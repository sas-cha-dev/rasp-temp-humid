import adafruit_dht
import board
import time
import redis
import json

# Redis connection
# Adjust host, port, and password as needed for your Redis instance
r = redis.Redis(host='localhost', port=6379, db=0, decode_responses=True)

# Initialisieren Sie den DHT, wobei der Datenpin mit Pin 16
# (GPIO 23) des Raspberry Pi verbunden ist:
dhtDevice = adafruit_dht.DHT22(board.D23)
dhtDevice2 = adafruit_dht.DHT22(board.D22)

while True:
    try:
        # Sensor 1 auslesen
        temperature_c = dhtDevice.temperature
        temperature_f = temperature_c * (9 / 5) + 32
        humidity = dhtDevice.humidity

        # Sensor 1 Daten zu Redis pushen
        sensor1_data = {
            'temperature_c': round(temperature_c, 1),
            'temperature_f': round(temperature_f, 1),
            'humidity': humidity,
            'timestamp': time.time()
        }
        r.set('sensor1', json.dumps(sensor1_data))

        print("Sensor1: {:.1f} F / {:.1f} C Humidity: {}%".format(
            temperature_f, temperature_c, humidity))

        # Sensor 2 auslesen
        temp_c2 = dhtDevice2.temperature
        temp_f2 = temp_c2 * (9/5) + 32
        humid2 = dhtDevice2.humidity

        # Sensor 2 Daten zu Redis pushen
        sensor2_data = {
            'temperature_c': round(temp_c2, 1),
            'temperature_f': round(temp_f2, 1),
            'humidity': humid2,
            'timestamp': time.time()
        }
        r.set('sensor2', json.dumps(sensor2_data))

        print("Sensor2: {:.1f} F / {:.1f} C Humidity: {}%".format(
            temp_f2, temp_c2, humid2))
        print("")

    except RuntimeError as error:
        # Fehler passieren ziemlich oft, DHT's sind schwer zu
        # lesen, einfach weitermachen
        print(error.args[0])
        time.sleep(2.0)
        continue
    except Exception as error:
        dhtDevice.exit()
        raise error

    time.sleep(4.0)
