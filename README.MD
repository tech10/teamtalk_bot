This is a bot for [TeamTalk](https://bearware.dk) that I started writing several years ago. I have not maintained it in some time, the code could probably be much better, and a lot of things could probably be changed, including the reliance on a readline library in Go that doesn't appear to be maintained any longer.

For the most part, the bot will guide you through any configuration needed. This software is incomplete, and I would consider any build alpha. As a result, it is subject to breaking changes.

I have provided builds with goreleaser in a GitHub action, and have started using this project to test some of my GitHub action features, such as cashing builds for quicker releases.

As it is, the program is stable, for the most part, though one person using it did report that sometimes the xml configuration file was created as a blank file, rather than configuration being written as it should be. I don't know why this is, and the issue, if it is still present, will be inconsistent, so I suggest you back up any configuration files.

Anyone is welcome to open issues, pull requests, and the like. I'll accept any contributions for this program, should they pass builds. I don't forsee actively maintaining this project, and if I do start actively maintaining it, I will likely be rewriting it in several different ways.

Good luck using this program if you are interested, and enjoy it. I hope it's useful to anyone using it.
