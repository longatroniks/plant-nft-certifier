import time
import threading

readings = []
last_message_time = None

# The MQTT callback thread appends while the main loop finalizes batches,
# so every access to the buffer goes through this lock.
lock = threading.Lock()

def add_reading(data):
    global last_message_time
    with lock:
        readings.append(data)
        last_message_time = time.time()
        return len(readings)

def get_last_message_time():
    return last_message_time

def take_batch():
    """Atomically hand over the buffered readings and start a new batch."""
    global readings, last_message_time
    with lock:
        batch = readings
        readings = []
        last_message_time = None
    return batch

def summarize_batch(batch):
    from statistics import mean

    def extract(field):
        return [r[field] for r in batch if field in r]

    def stats(field):
        values = extract(field)
        return {
            "avg": round(mean(values), 2),
            "min": round(min(values), 2),
            "max": round(max(values), 2),
        }

    summary = {
        "start_timestamp": batch[0]["timestamp"],
        "end_timestamp": batch[-1]["timestamp"],
        "reading_count": len(batch),
        "summary": {
            "temperature": stats("temperature"),
            "humidity": stats("humidity"),
            "pressure": stats("pressure"),
        }
    }
    return summary
