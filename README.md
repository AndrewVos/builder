# builder

Very simple CI

## Features
  * Auto builds new pushes and pull requests to Github
  * Run builds with Github hook
  * Display a list of builds
  * Clicking on a build displays the build output, with full colour

## Usage

[Get a Github Auth Token](https://help.github.com/articles/creating-an-access-token-for-command-line-use)

Set the following environment variables:

		GITHUB_CLIENT_ID=
		GITHUB_CLIENT_SECRET=
		HOST=
		PORT=

Host will be the static IP or hostname of the server that builder is running on.

Repositories is a list of repositories you want watched.

Launch builder:

    go build
    ./builder

Add a ``Builderfile`` to your projects that you want to build.
A typical Builderfile looks something like this:

    #!/bin/bash

    make test # or some other sort of test runner thingy

Go to host:port to view a list of builds

## Hooks

Hooks get executed whenever a build completes. To add a new hook just save a script in ```data/hooks```.

These are the available environment variables:

      $BUILDER_BUILD_RESULT # pass, fail or incomplete
      $BUILDER_BUILD_URL    # the build url
      $BUILDER_BUILD_ID     # unique build ID
      $BUILDER_BUILD_OWNER  # username of commit owner
      $BUILDER_BUILD_REPO   # repository name
      $BUILDER_BUILD_REF    # branch name
      $BUILDER_BUILD_SHA    # commit SHA
