
# HTTP From TCP

Implementation of the HTTP/1.1 server (limited) from scratch leaning on the relevant RFC's as well as the `net/http` library in Golang.

## Installation and Running

To run the server you will need Golang. I used 1.25.1, but I'm quite sure it's backwards compatible (but not too backwards). <a href="https://go.dev/doc/install">Install Golang here</a>.

Each commit (for better or worse) represents a lesson from boot.dev course below. Note that the last lesson titled "Binary Data" requires to complete the curl to pull the video.

Run the server (assuming the `pwd` is root) with:
`go run cmd/httpserver/main.go`

## Notes About the Code

Almost all of the code was written without AI (apart from some menial tasks in certain lessons). As you can see I'm not a fan of the seemingly idomatic way of Go's declarations using `:=`, so I don't use it at all. The tests however were written to major extend by AI (Codex specifically), thus `:=` is in constant use there.

## Highlights

- Plain TCP handling with custom request/response types
- Chunked proxy responses with trailer hashing
- Simple static routes, including a demo MP4 streamer

## License

Distributed under the MIT License. See [`LICENSE`](./LICENSE) for details.

## Learn More

- <a href="https://www.boot.dev/lessons/b0cebf37-7151-48db-ad8a-0f9399f94c58">Boot.dev: Build HTTP/1.1 from Scratch</a>
- <a href="https://www.youtube.com/watch?v=FknTw9bJsXM">ThePrimeagen: Video Walkthrough</a>
