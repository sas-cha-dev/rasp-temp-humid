# Temperature & Humidity Logger

A Go-based temperature, humidity, button push monitoring system for Raspberry Pi with DHT22 sensors.

The code is mainly build by claude.ai to get the system up fast and running. Some adjustments were made.

## Project Structure

```
temp-humidity-logger/
├── cmd/
│   └── server/
│       └── main.go              # Main application entry point
├── internal/
│   ├── sensor/
│   │   ├── sensor.go            # Sensor interface
│   │   └── dummy.go             # Dummy implementation (for testing)
│   ├── repository/
│   │   └── sqlite.go            # SQLite database layer
│   └── handler/
│       └── handler.go           # HTTP handlers
├── web/
│   └── index.html               # Dashboard UI
├── go.mod
├── .env                         # Environment configuration
└── README.md
```

## Features

- ✅ Abstract sensor interface with dummy implementation
- ✅ Dummy service returns values that change every second (sine wave simulation)
- ✅ SQLite database with repository pattern
- ✅ Configurable reading interval via environment variable
- ✅ Dashboard displaying:
    - Latest readings from both sensors with timestamp
    - Average values for last hour, today, and this week
    - Table of last 100 readings
- ✅ Auto-refreshing UI (every 30 seconds)

## Setup

### 1. Install Dependencies

```bash
go mod download
```

### 2. Configure Environment

Edit the `.env` file or set environment variables:

```bash
READ_INTERVAL=10    # Read sensors every 10 seconds
DB_PATH=./data.db   # Database file path
PORT=8080           # HTTP server port
TEMPLATE_DIR=./web  # Template directory
```

### 3. Run the Server

```bash
# Load environment variables and run
source .env
go run cmd/server/main.go
```

Or build and run:

```bash
go build -o temp-logger cmd/server/main.go
./temp-logger
```

### 4. Access Dashboard

Open your browser and navigate to:
```
http://localhost:8080
```

## API Endpoints

- `GET /` - Main dashboard (HTML)
- `GET /api/data` - JSON API endpoint with all data

## Dummy Sensor Behavior

The dummy sensor simulates realistic temperature and humidity readings:

- **Sensor 1**: Temperature oscillates between 20-24°C, Humidity between 45-55%
- **Sensor 2**: Temperature oscillates between 19-23°C, Humidity between 50-60%
- Values change based on sine/cosine waves over time
- Updates continuously as the application runs

## Database Schema

```sql
CREATE TABLE readings (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    sensor_id INTEGER NOT NULL,
    temperature REAL NOT NULL,
    humidity REAL NOT NULL,
    timestamp DATETIME NOT NULL
);
```

## OneDrive Backup

For OneDrive backup, you can use `rclone`:

```bash
# Install rclone
curl https://rclone.org/install.sh | sudo bash

# Configure OneDrive
rclone config

# Create a backup script
cat > backup.sh << 'EOF'
#!/bin/bash
rclone copy ./data.db onedrive:temp-humidity-backup/
EOF

chmod +x backup.sh

# Add to crontab (daily backup at 2 AM)
crontab -e
# Add: 0 2 * * * /path/to/backup.sh
```

## Development

### Running Tests

```bash
go test ./...
```

### Viewing Logs

The application logs sensor readings and any errors to stdout:

```
2024-11-17 15:04:23 Saved: Sensor 1 - Temp: 22.5°C, Humidity: 52.3%, Time: 15:04:23
2024-11-17 15:04:23 Saved: Sensor 2 - Temp: 21.8°C, Humidity: 56.1%, Time: 15:04:23
```

## Troubleshooting

**Database locked error**: Make sure only one instance is running

**No data showing**: Wait for the first reading interval to pass (default 10 seconds)

**Port already in use**: Change the `PORT` environment variable

## License

MIT