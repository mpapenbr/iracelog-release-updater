go install github.com/cespare/reflex@latest
go install github.com/goreleaser/goreleaser@latest
go install github.com/caarlos0/svu@latest

if [ -f setuplinks.sh ]; then
    . ./setuplinks.sh
fi