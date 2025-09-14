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
- `REDIS_ADDR`: The address of the Redis server (default: `localhost:6379`).

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

3. Set up a Redis server if not already running. You can use Docker to quickly set up Redis:
   ```bash
   docker run --name redis -d -p 6379:6379 redis
   ```

4. Set the required environment variables:
   ```bash
   export DISCORD_WEBHOOK_URL=<your-discord-webhook-url>
   export REDIS_ADDR=localhost:6379
   ```

## Usage
Run the application:
```bash
go run main.go
```

## How It Works
1. The application fetches the latest posts from the /r/borderlandsshiftcodes subreddit.
2. It extracts shift codes using a regex pattern.
3. It checks Redis to filter out already processed codes.
4. New codes are sent to the specified Discord webhook.

## Contributing
Feel free to submit issues or pull requests if you find any bugs or have suggestions for improvements.

## License
This project is licensed under the MIT License.
