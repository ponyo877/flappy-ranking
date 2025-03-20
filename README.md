# Flappy Gopher With Ranking

## Overview

This is a version of the Flappy Gopher game with added ranking functionality. Players can play the game and submit their scores to the server to participate in the rankings.

## Technology Stack

### Client
- [Ebiten](https://ebiten.org/) - A 2D game library written in Go

### Server
- [Cloudflare Workers](https://workers.cloudflare.com/) - Serverless platform
- [Cloudflare D1](https://developers.cloudflare.com/d1/) - SQLite-based serverless database
- [github.com/syumai/workers](https://github.com/syumai/workers) - Library for developing Cloudflare Workers in Go

## Setup

### Prerequisites
- Go 1.24 or higher
- npm and Node.js
- Cloudflare account
- Wrangler CLI (`npm install wrangler --save-dev`)

### Creating a Cloudflare D1 Database

1. Create a D1 database using the Wrangler CLI:

```bash
wrangler d1 create flappy_ranking
```

2. Set the created database ID in the `database_id` field of the wrangler.toml file.

3. Apply the schema:

```bash
wrangler d1 execute flappy_ranking --file ./storage/d1/schema.sql
```

### Build and Deploy

1. Build and deploy the Workers application:

```bash
wrangler deploy
```

2. After deployment, set the URL displayed in the Cloudflare Dashboard to the `serverEndpoint` variable in client/main.go.

### Running the Client

```bash
go run ./client
```

## Development

### Local Development

To run Workers locally:

```bash
wrangler dev
```

To run the client locally, set the `serverEndpoint` variable in `client/main.go` to your local Workers endpoint (usually `http://localhost:8080`).

## License

This project is licensed under the Apache License 2.0. See the LICENSE file for details.

## Credits

- The Go Gopher character was designed by [Renee French](https://reneefrench.blogspot.com/) and is licensed under [CC BY 3.0](https://creativecommons.org/licenses/by/3.0/).
- The original Flappy Gopher game was developed by the [Ebiten](https://ebiten.org/) team.
