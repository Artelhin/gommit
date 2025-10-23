package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

var (
	mFlag   string
	preFlag string
	sufFlag string
)

var configFile *os.File

const defaultConfig = `{
	"branches": {}
}`

func init() {
	flag.StringVar(&mFlag, "m", "", "-m <msg> use given message as git commit -m <msg> with configured pre- and suffixes")
	flag.StringVar(&preFlag, "pre", "", "-pre <prefix> use given prefix as a default prefix for this branch")
	flag.StringVar(&sufFlag, "suf", "", "-suf <sufffix> use given sufffix as a default sufffix for this branch")
	flag.Parse()
}

type Config struct {
	Branches map[string]BranchConfig `json:"branches"`
}

type BranchConfig struct {
	Prefix string `json:"prefix"`
	Suffix string `json:"suffix"`
}

func main() {
	config, err := parseBranchPrefixMappingFile()
	if err != nil {
		fmt.Println(err)
		return
	}

	cmd := exec.Command("git", "branch", "--show-current")
	out, err := cmd.Output()
	branch := string(out[:len(out)-1]) // drop /n symbol
	bconfig := config.Branches[branch]

	var (
		prefix, suffix string
		updateNeeded   bool
	)

	if preFlag != "" {
		updateNeeded = true
		prefix = preFlag
	} else {
		prefix = bconfig.Prefix
	}
	if sufFlag != "" {
		updateNeeded = true
		suffix = sufFlag
	} else {
		suffix = bconfig.Suffix
	}

	if mFlag == "" {
		fmt.Println("abort due to empty commit message")
		return
	}

	message := strings.Join([]string{prefix, mFlag, suffix}, " ")
	fmt.Printf("commiting to branch ->> %s\n", branch)

	cmd = exec.Command("git", "commit", "-m", message)
	out, err = cmd.Output()
	if err != nil {
		fmt.Printf("can't exec git commit: %s\n", err)
	}
	fmt.Print(string(out))

	if updateNeeded {
		bconfig.Prefix = prefix
		bconfig.Suffix = suffix
		config.Branches[branch] = bconfig
		b, err := json.Marshal(config)
		if err != nil {
			fmt.Printf("failed to update config: %s\nedit manually at .git/gommit.json", err)
			return
		}
		err = modifyConfig(b)
		if err != nil {
			fmt.Printf("failed to write to config file: %s\nedit manually at .git/gommit.json", err)
		}
	}
}

func parseBranchPrefixMappingFile() (*Config, error) {
	gitPath, err := findGitDir()
	if err != nil {
		return nil, fmt.Errorf("can't find .git dir: %s", err)
	}

	configFile, err = os.OpenFile(gitPath+"/gommit.json", os.O_CREATE, os.ModePerm)
	b, err := io.ReadAll(configFile)
	if err != nil {
		return nil, fmt.Errorf("can't read gommit config: %s", err)
	}

	if len(b) == 0 {
		b = []byte(defaultConfig)
	}

	var config Config
	err = json.Unmarshal(b, &config)
	if err != nil {
		return nil, fmt.Errorf("can't parse gommit config: %s", err)
	}
	return &config, nil
}

func findGitDir() (string, error) {
	cdUpRegexp := regexp.MustCompile(`\/[^\/]*$`)
	curDir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("can't get current directory: %s", err)
	}
	for curDir != "" {
		stat, err := os.Stat(".git")
		if err != nil && !errors.Is(err, fs.ErrNotExist) {
			return "", fmt.Errorf("can't look for .git directory: %s", err)
		}
		if err == nil {
			if stat.IsDir() {
				return curDir + "/.git", nil
			}
		}
		curDir = cdUpRegexp.ReplaceAllString(curDir, "")
		if err = os.Chdir(curDir); err != nil {
			return "", fmt.Errorf("can't change working directory: %s", err)
		}
	}
	return "", errors.New("not a Git repository")
}

func modifyConfig(b []byte) error {
	_, err := configFile.WriteAt(b, 0)
	return err
}
