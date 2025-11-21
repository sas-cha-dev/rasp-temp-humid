import time
import board
import adafruit_dht

# Initialisieren Sie den DHT, wobei der Datenpin mit Pin 16
# (GPIO 23) des Raspberry Pi verbunden ist:
dhtDevice = adafruit_dht.DHT22(board.D23)
dhtDevice2 = adafruit_dht.DHT22(board.D22)
while True:
    try:
        # Ausgabe der Werte über die serielle Schnittstelle
        temperature_c = dhtDevice.temperature
        temperature_f = temperature_c * (9 / 5) + 32
        humidity = dhtDevice.humidity
        print("Sensor1: {:.1f} F / {:.1f} C Humidity: {}%".format(temperature_f, temperature_c, humidity))
        temp_c2 = dhtDevice2.temperature
        temp_f2 = temp_c2 * (9/5) + 32
        humid2 = dhtDevice2.humidity
        print("Sensor2: {:.1f} F / {:.1f} C Humidity: {}%".format(temp_f2, temp_c2, humid2))
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