data
====

## 20150622-hd2001-ramp-up.pdf

spikes probably coming from too fast queries on the CANBus coming from:
```java
@Override
public void tick() {
    try {
		readTemperature();
		readPressure();
		readHygrometry();
		this.publish("temperature", this.temperature);
		this.publish("hygrometry",  this.hygrometry);
		this.publish("pressure",    this.pressure);
	} catch (FcsHardwareException) { ... }
}
```

## 20150624-hd2001-3sensors-200ms-sleep.pdf

modulation coming from ``sleeps`` in:
```java
@Override
public void tick() {
    try {
		readTemperature();
		try {
			Thread.sleep(200);
		} catch(InterruptedException) {
			Thread.currentThread().interrupt();
		}
		readPressure();

		try {
			Thread.sleep(200);
		} catch(InterruptedException) {
			Thread.currentThread().interrupt();
		}
		readHygrometry();
		this.publish("temperature", this.temperature);
		this.publish("hygrometry",  this.hygrometry);
		this.publish("pressure",    this.pressure);
	} catch (FcsHardwareException) { ... }
}
```
