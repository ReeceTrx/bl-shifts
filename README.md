# Borderlands Shift Codes Fetcher

This application fetches the latest Borderlands shift codes from the subreddit [r/borderlandsshiftcodes](https://www.reddit.com/r/borderlandsshiftcodes) and sends them to a Discord webhook. It uses Redis to store and filter out already processed codes to avoid duplicate notifications.

## Features
- Fetches the latest posts from the subreddit.
- Extracts valid shift codes using regex.
- Stores processed codes in Redis to prevent duplicates.
- Sends new shift codes to a Discord channel via a webhook.

## Requirements

### Environment Variables
The application requires the following environment variables to be set:

- `DISCORD_WEBHOOK_URL`: The Discord webhook URL where the shift codes will be sent. (*If not set, no notifications will be sent*)
- `REDIS_ADDR`: The address of the Redis server (e.g., `localhost:6379`).
- `FILENAME`: The path to the file where seen codes will be stored. (*If not set, defaults to `codes.json` if `REDIS_ADDR` is not provided*)
- `INTERVAL_MINUTES`: The interval in minutes between checks. (*If not set, the application runs once and exits*)

### Discord Webhook URL
>[!DANGER] Do not share your Discord Webhook with anyone, as it can be used to post messages into your Discord server without requiring authentication

To get your Discord Webhook URL, you can follow the instructions in [Discords Website](https://support.discord.com/hc/en-us/articles/228383668-Intro-to-Webhooks).

## Installation
1. Clone the repository:
   ```bash
   git clone github.com/imdevinc/bl-shifts
   cd bl-shifts
   ```

2. Install dependencies:
   ```bash
   go mod tidy
   ```

3. Set the required environment variables (if applicable):
   ```bash
   export DISCORD_WEBHOOK_URL=<your-discord-webhook-url>
   export FILENAME=codes.json
   export INTERVAL_MINUTES=15
   ```

## Usage

### Running Locally
Run the application:
```bash
go run main.go
```

### Running with Docker
You can also run the application using the pre-built Docker image:
```bash
docker run --rm \
  -e DISCORD_WEBHOOK_URL=<your-discord-webhook-url> \
  -e REDIS_ADDR=localhost:6379 \
  ghcr.io/imdevinc/bl-shifts:latest
```

## How It Works
1. The application fetches the latest posts from the /r/borderlandsshiftcodes subreddit.
2. It extracts shift codes using a regex pattern.
3. It checks the configured storage (Redis or file) to filter out already processed codes.
4. New codes are sent to the specified Discord webhook (if configured).

## Contributing
Feel free to submit issues or pull requests if you find any bugs or have suggestions for improvements.

## License
This project is licensed under the MIT License.
