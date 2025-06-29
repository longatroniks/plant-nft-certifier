import time

readings = []
last_message_time = None

def update_last_message_time():
    global last_message_time
    last_message_time = time.time()

def get_last_message_time():
    return last_message_time

def reset():
    global readings, last_message_time
    readings.clear()
    last_message_time = None

def summarize_batch(batch):
    from statistics import mean

    def extract(field):
        return [r[field] for r in batch if field in r]

    summary = {
        "start_timestamp": batch[0]["timestamp"],
        "end_timestamp": batch[-1]["timestamp"],
        "summary": {
            "temperature": {
                "avg": round(mean(extract("temperature")), 2),
                "min": round(min(extract("temperature")), 2),
                "max": round(max(extract("temperature")), 2),
            },
            "humidity": {
                "avg": round(mean(extract("humidity")), 2),
                "min": round(min(extract("humidity")), 2),
                "max": round(max(extract("humidity")), 2),
            },
            "pressure": {
                "avg": round(mean(extract("pressure")), 2),
                "min": round(min(extract("pressure")), 2),
                "max": round(max(extract("pressure")), 2),
            },
        }
    }
    return summary
