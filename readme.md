CLI 

gitfresh init

requesting

- confirm working directory
- entry NGROK TOKEN
- entry NGROK CUSTOM DOMAIN (optional)
- entry GITHUB TOKEN


Step Step

gitfresh config

gitfresh init  

- Scan the GitWorkDir to discovery the git repositories
- Create repositories.json with all available repos in GitWorkDir
- Start the local webhook-listener server
- Create webhook integration on github.com for each available repository.

gitfresh status

list all repositories with your status

repo=app-ngx status=synced updated_date=20/12/11 20:12 msg="successfully updated"
repo=app-corin status=outdated updated_date=20/12/11 20:12 msg="failed when try update"