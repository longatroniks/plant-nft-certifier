def on_connect(client, userdata, flags, rc):
    print(f"[MQTT] Connected with result code {rc}")
    client.subscribe(userdata["topic"])

def on_message(client, userdata, msg):
    from collector.data_collector import readings, update_last_message_time
    import json

    try:
        data = json.loads(msg.payload.decode())
        if all(k in data for k in ("temperature", "humidity", "pressure", "timestamp")):
            readings.append(data)
            update_last_message_time()
            print(f"[+] Collected reading #{len(readings)}")
    except Exception as e:
        print(f"[!] Error processing message: {e}")
