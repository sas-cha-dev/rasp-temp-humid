# Raspberry Pi Monitoring System

A **Go-based** monitoring system for temperature, humidity, button inputs, and weather data running on a Raspberry Pi.

The initial code was generated in a claude.ai chat to quickly get the system up and running, although not much of the 
original code remains.

My setup uses a **Raspberry Pi 3** running **Debian GNU/Linux 13 (trixie)**.

Two **DHT22** sensors are connected to the GPIO pins **GPIO 22 (pin 15)** and **GPIO 23 (pin 16)**. 
They communicate via a one-wire interface, which may need to be enabled on the Raspberry Pi beforehand.

To read sensor values, a small **Python script** is used, since reading DHT22 data directly from Go can be challenging. 
The script reads values every **4 seconds** (minimum interval: **2 seconds**) and writes them to a **Redis** database 
using the keys `sensor1` and `sensor2`.

The script can be run persistently using `systemctl` and a corresponding service definition.

A button is used to represent a ventilation event. The button is read inside the Go application using the **rpio** library.

Additionally, weather data (temperature and humidity) is fetched from **OpenWeather**.

Each module—DHT sensors, button, and weather—is independent. All collected data is logged into a **SQLite** database.

I hope this project is helpful to others. If you make improvements, feel free to share them! For new features, it would 
be great if they can be enabled via environment variables or command-line parameters.

---

## Table of Contents

* [Environment Variables](#environment-variables)
* [API Endpoints](#api-endpoints)
* [Development Environment](#development-environment)
* [Troubleshooting](#troubleshooting)
* [Example `.env`](#example-env)
* [Example systemd Service](#example-systemd-service)

---

## Environment Variables

| Variable                    | Description                                                            |
| --------------------------- | ---------------------------------------------------------------------- |
| `READ_INTERVAL`             | Interval in seconds for reading the DHT sensors                        |
| `DB_PATH`                   | Path to the SQLite database (may be relative to the executable)        |
| `PORT`                      | Port for the web server                                                |
| `TEMPLATE_DIR`              | Directory containing the HTML templates                                |
| `WEATHER_READ_INTERVAL_MIN` | Interval in minutes for requesting data from OpenWeather               |
| `OPEN_WEATHER_API_KEY`      | API key for the OpenWeather OneCall endpoint                           |
| `LOCATION_COORDS`           | Latitude and longitude for the OpenWeather request (format: `lat,lon`) |

---

## API Endpoints

* **GET /** — Main dashboard (HTML)
* **GET /api/data** — JSON API endpoint containing all collected data

---

## Development Environment

To avoid developing directly on the Raspberry Pi, the project includes separate entry points (see `/cmd/`) as well as 
dummy services for the DHT sensors and the button. These mock services behave randomly to simulate real-world input.

---

## Troubleshooting

**Database locked error**

Make sure that only a single instance of the application is running.

**Port already in use**

Change the `PORT` environment variable to an available port.

---

## Example `.env`

```env
# Read interval in seconds for DHT sensors (default: 4)
READ_INTERVAL=4

# SQLite DB path
DB_PATH=./data/sensor_data.db

# Web server port
PORT=8080

# HTML templates directory
TEMPLATE_DIR=./templates

# Weather polling interval (minutes)
WEATHER_READ_INTERVAL_MIN=15

# OpenWeather API key and coordinates
OPEN_WEATHER_API_KEY=your_api_key_here
LOCATION_COORDS=48.1371,11.5754
```

---

## Example systemd Service

Create a systemd file (e.g. `/etc/systemd/system/dht-reader.service`) for the Python sensor reader:

```ini
[Unit]
Description=DHT22 Sensors to Redis
After=network.target redis.service

[Service]
Type=simple
User=<username>
WorkingDirectory=/repo/SensorsPython
ExecStart=/repo/SensorsPython/env/bin/python /repo/SensorsPython/dht22.py
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
```

Enable and start with:

```bash
sudo systemctl daemon-reload
sudo systemctl enable dht-reader.service
sudo systemctl start dht-reader.service
```

---