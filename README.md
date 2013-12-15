# builder

Very simple CI

## Features
  * Auto builds new pushes and pull requests to Github
  * Run builds with Github hook
  * Display a list of builds
  * Clicking on a build displays the build output

## Usage

Setup a a builder.json file:

    {
      "AuthToken": ""
      "Host": "",
      "Port": "",
      "Repositories": [
        {"Owner": "", "Repository": ""}
      ]
    }

[How to create an auth token](https://help.github.com/articles/creating-an-access-token-for-command-line-use).

Host will be the static IP or hostname of the server that builder is running on.

Repositories is a list of repositories you want to build.

Launch builder:

    go build
    ./builder

Add a ``Builderfile`` to your projects that you want to build.
A typical Builderfile looks something like this:

    #!/bin/bash

    make test # or some other sort of test runner thingy
