```zsh
brew install hugo
go install github.com/gohugoio/hugo@latest
CGO_ENABLED=1 go install -tags extended,withdeploy github.com/gohugoio/hugo@latest
hugo new site blog
hugo new theme blog
hugo server
```
