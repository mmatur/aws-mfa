FROM golang:1.11

LABEL "name"="golang tools action"
LABEL "maintainer"="Michael Matur <michael.matur@gmail.com>"
LABEL "version"="1.0.0"

LABEL "com.github.actions.icon"="code"
LABEL "com.github.actions.color"="green-dark"
LABEL "com.github.actions.name"="Go Tools"
LABEL "com.github.actions.description"="This is an Action to run go."

RUN curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | sh -s -- -b $GOPATH/bin v1.12.3

COPY entrypoint.sh /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]