[![Coverage
Status](https://coveralls.io/repos/github/AndreasSko/go-library-merger/badge.svg?branch=master)](https://coveralls.io/github/AndreasSko/go-library-merger?branch=master)

# library-merger
A command-line tool to easily merge JW Library® backups, written in Go.
For the iOS version, visit [ios-jwlm](https://github.com/AndreasSko/ios-jwlm).

library-merger allows you to merge two .jwlibrary backup files, while giving you
control of the process - your notes are precious, and you shouldn‘t need to
trust a program solving possible merge conflicts for you.

I created this project with the goal of having a tool that is able to work on
multiple operating systems, and even allowing it to be incorporated in other
programs as a library (like an [iOS app](https://github.com/AndreasSko/ios-jwlm)).
It is - and will be for
quite some time - a work-in-progress project, so I‘m always open for suggestion
and especially reports if you encounter an unexpected behavior or other bugs. 

The usage is pretty simple: you have one command, you name your backup files -
and press enter. The tool will merge all entries for you. If it encounters a
conflict (like the same note with different content or two markings that
overlap), it will ask you for directions: should it choose the left version or
the right one? After that is finished, you have a nicely merged backup that you
can import into your JW Library® App. The first merge process might take some
time because of the number of possible conflicts, depending on how far apart you
backups are. But if you merge them regularly, it should be a matter of seconds
:) 

## Usage
```shell
library-merger merge <left-backup> <right-backup> <merged-backup>
```

If a conflict occurs while merging, the tool will ask for directions: should it
choose the left version or the right one. For that, it shows you the actual
entries (I‘m planning to improve that view and add more information, especially
about publications, in the future). If you are not sure what to do, press `?`
for help. 

### Resolve conflicts automatically
Currently, there are three solvers you can use to automatically resolve
conflicts: `chooseLeft`, `chooseRight`, and `chooseNewest` (though the last one
is only usable for Notes). As their names suggest, `chooseLeft` and
`chooseRight` will always choose the same side if a conflict occurs, while
`chooseNewest` always chooses the newest entry. 

You can enable these solvers with the `--bookmarks`, `--markings`,
`--notes`, and `--inputFields` flags:

```shell
library-merger merge <left-backup> <right-backup> <merged-backup> --bookmarks chooseLeft --markings chooseRight --notes chooseNewest --inputFields chooseLeft
```

The conflict resolvers are helpful for regular merging when you are 
sure that one side is always the most up-to-date one. For a first merge, 
it is still recommended to manually solve conflicts, so you don't risk
accidentally overwriting entries.

### Compare two backups
To quickly compare two backup files and check if their content is equal,
you can use the `library-merger compare <left-backup> <right-backup>` command. 
This one is mainly used for validation, but might be helpful in other 
situations :)

## Installation 
You can find the compiled binaries for Windows, Linux, and Mac under the
[Release](https://github.com/AndreasSko/go-library-merger/releases) section. 

### Installation using Homebrew (Mac and Linux)
library-merger can be easily installed using Homebrew:
```shell
brew install andreassko/homebrew-go-library-merger/library-merger
```

See the instructions on how to install Homebrew at https://brew.sh

## Mobile version
If you want to merge backups using your iPhone or iPad, have a look at
[Library Merger](https://github.com/AndreasSko/ios-jwlm). It uses the whole merge
logic of library-merger, but wraps it in a nice and easy to use iOS app. It is 
already available on the [App Store](https://apps.apple.com/us/app/jwlm-jw-library-merger/id1539780103).

There might come an Android version at some point, but as I personally
don't use any Android devices, it is - unfortunately - not the highest
priority for me right now. If you are interested in helping with this
project, feel free to contact me or just start it on your own :)

### Technical background
[Gomobile](https://github.com/golang/go/wiki/Mobile) enables to 
include Go packages within a mobile application. As there are still some 
[limitations](https://pkg.go.dev/golang.org/x/mobile/cmd/gobind#hdr-Type_restrictions),
I built a wrapper for most structs and methods. To build the binding
on your own, install Gomobile, change into the `gomobile` directory
of this repo and run `gomobile bind -target <ios or android>`. 

## A word of caution 
It took me a while to trust my own program, but I still keep backups of my
libraries - and so should you. Library-merger is still in beta-phase, so there is a
possibility that something might get lost because of a yet-to-find bug. So
please keep that in mind and - again - if you found a bug, feel free to open an
issue. 

## Need help?
Something is unclear, you have suggestions for documentation or you found a bug?
Feel free to open an issue. I‘m happy to help, though please be patient if it
takes a while for me to respond :)

## Legal Notice

This application is not owned or operated by
Watch Tower Bible and Tract Society of Pennsylvania, however it is distributed
with permission under a license from Watch Tower. All ownership rights to the
content within the JW Library® application remain exclusively with Watch Tower
Bible and Tract Society of Pennsylvania.

JW Library® is a registered trademark owned exclusively by Watch Tower Bible
and Tract Society of Pennsylvania ("Watch Tower").
