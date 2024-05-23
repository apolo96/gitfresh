# GitFresh
A Developer Experience Tool to Keep Git local Repositories Updated 😎

## Use Cases

**Collaborative Software Development**: In software development teams, multiple members may work on a shared project. Keeping local repositories updated ensures that everyone is working with the latest version of the code and reduces integration conflicts.

**Frontend and Backend Integration**: In a local integration environment, a frontend developer needs to keep their local repository updated with the latest changes from the develop backend API, which is being developed by a teammate, to ensure smooth integration.

> Feel free to use the gitfresh tool for other scenarios as well!

## Install

Gitfresh is available for MacOS and Linux via Homebrew.

First, you should install [HomeBrew](https://brew.sh/).

Then, run the following command:

```bash
brew install apolo96/tap/gitfresh
```

Check installation:

```bash
gitfresh version
```

## Quickstart

GitFresh is a tool powered by Github and Ngrok.

> I hope to add support to other git server providers in the future.

### Requirements

- Github Token
- Ngrok Token

You can go to Github to create a new token with `admin:repo_hook` scope: 

https://github.com/settings/tokens

You can go to Ngrok to get a personal token:

https://dashboard.ngrok.com/get-started/your-authtoken

### Initialize Workspace

After installing GitFresh via Homebrew and getting GitHub and Ngrok tokens, then you can initialize a workspace:

First, go to the working directory that contains the Git repositories and run the following command:

```bash
gitfresh config
```

It opens the CLI Form where you should enter the GitHub and Ngrok tokens.

The TunnelDomain input is OPTIONAL, by default Ngrok generates a random DNS.

If you prefer, you can create a custom domain at: 
https://dashboard.ngrok.com/cloud-edge/domains.


Finally, run the following command:

```bash
gitfresh init
```

This command can take some seconds for startup services. 

### Add new repository

You can add new repositories to Gitfresh. Running the following command:

```bash
gitfresh scan
```

### Discover the CLI

```bash
gitfresh -help
```


## How It Works

GitFresh creates GitHub webhooks to send notifications of events git-push through an internet tunnel provided by Ngrok that triggers repository updates on the local machine (gitfresh agent)

![Imagen de un gato](https://i.ibb.co/R6kPhmW/gitfresh.png)
