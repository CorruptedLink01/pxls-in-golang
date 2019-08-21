# pxls.space in Golang

Rework of the [pxls.space](https://pxls.space) (a r/Place clone) backend in Golang.

This started from the frustration of getting the original, Java version, to work; and was impulsed by other Pxls' Developers talking about a rework in either Golang or Rust.

I'm fairly new to Rust and had learned Golang a few months ago, so prefered trying to refresh what I knew about Golang instead.

With some hope (and luck), this might eventually become the live version of the game.

## Development
### Setup
1. Install [Golang 1.12.9](https://golang.org/) or later
2. Clone this repository and `cd` to it
3. Run `go mod download` to install dependencies
4. Clone `reference.pxls.conf`, rename it to `pxls.conf` and configure to your liking

### For running
5. Run `go run src/*`
6. Go to [https://localhost:4567](https://localhost:4567) (or whatever port you configured in `pxls.conf`)


## Implemented
- [x] Load pxls.conf
- [ ] Load boarddata.bin
- [ ] Webserver
	- [x] Sending static content to the client
	- [x] /info endpoint
	- [x] /boarddata endpoint
	- [x] /whoami endpoint
	- [ ] Oauth endpoints
	- [ ] other endpoints...
- [ ] Websocket
	- [x] send userinfo message
		- [x] IP-based
		- [x] Token-based
	- [x] handle pixel placing
	- [x] handle pixel stacking
	- [ ] other message types...
- [ ] User roles
- [ ] Database
- [ ] Console commands


## Plans
- [ ] Separate frontend-serving code from game-handling backend code

### Ambitious ideas
- [ ] Tools
	- [ ] /stats in Golang
	- [ ] Database overview for moderators in Golang
	- [ ] boarddata.bin maker
- [ ] Frontend
	- [ ] Modularize/[Webpacketize](https://github.com/go-webpack/webpack)
	- [ ] Convert to [TypeScript](https://www.typescriptlang.org/)
