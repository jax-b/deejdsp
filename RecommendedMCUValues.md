# Recommended delay and baud settings

Slower micro controllers need some time to do their stuff. Here is the recommendation that i played with for the MCU's i have on hand

| MCU               | Arduino Names | Baud Rate | Startup Delay | Command Delay |
|------------------:|:--------------|:---------:|:-------------:|:-------------:|
| Mega328P          | Nano          | 57600     | 3500          | 45            |
| Mega32u4 (16MGHz) | Micro         | 115200    | 0             | 0             |
