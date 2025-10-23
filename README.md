# gommit

<b>gommit</b> is a tool for more dynamic Git commit prefixes than Git commit prehooks.
It expands on the most useful prehook functionality in my opinion which is adding <b>branch name</b> and <b>service name</b> in case of several microservices existing in the same Git repository.

## Installation
```
go install github.com/Artelhin/gommit
```

## Usage
To use gommit, simply type 
```
gommit -m <commit message>
```
It will automatically create ```gommit.json``` config file in your .git repository.
By default it's empty, but you can edit it either manually or by using ```-pre``` or ```-suf``` flags to assign default prefix and suffix respectively.

Example:

```
gommit -pre 'QUEUE-123' -suf '[skip-ci]' -m 'refactor documentation'
```

## Configuration

```json gommit.json
{
    "branches": {
        "master": {
            "prefix": "[master]"
        },
        "dev": {
            "prefix": "[develop]",
            "suffix": "[skip-linter]"
        },
        "QUEUE-123": {
            "prefix": "QUEUE-123",
        }
    }
}
```