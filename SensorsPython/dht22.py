import time
import board
import adafruit_dht

# Initialisieren Sie den DHT, wobei der Datenpin mit Pin 16
# (GPIO 23) des Raspberry Pi verbunden ist:
dhtDevice = adafruit_dht.DHT22(board.D23)
while True:
    try:
# Ausgabe der Werte über die serielle Schnittstelle
        temperature_c = dhtDevice.temperature
        temperature_f = temperature_c * (9 / 5) + 32
        humidity = dhtDevice.humidity
        print("Temp: {:.1f} F / {:.1f} C Luftfeuchtigkeit: {}%".format(temperature_f, temperature_c, humidity))
    except RuntimeError as error:
# Fehler passieren ziemlich oft, DHT's sind schwer zu
# lesen, einfach weitermachen
        print(error.args[0])
        time.sleep(2.0)
        continue
    except Exception as error:
        dhtDevice.exit()
    raise error
time.sleep(2.0)