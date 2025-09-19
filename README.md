
# HTTP From TCP

Implementation of the HTTP/1.1 server (limited) from scratch leaning on the relevant RFC's as well as the `net/http` library in Golang. Each commit is based on an assignment from the boot.dev lessons, and implements itteratively the HTTP parser and the rest of the protocol features.

## Highlights

- Plain TCP handling with custom request/response types
- Chunked proxy responses with trailer hashing
- Simple static routes, including a demo MP4 streamer

## Learn More

- <a href="https://www.boot.dev/lessons/b0cebf37-7151-48db-ad8a-0f9399f94c58">Boot.dev: Build HTTP/1.1 from Scratch</a>
- <a href="https://www.youtube.com/watch?v=FknTw9bJsXM">ThePrimeagen: Video Walkthrough</a>
