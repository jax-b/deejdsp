# Arduino Deej Base Code
## The Basis for all modules to be built

#### Here's a summary

The Arduino is constantly listening for data to be sent to its RX port When it receives a byte it then starts to look for a newline character. The newline character marks the end of a command. After finding the new line it looks through its known command list and if it finds a match execute the appropriate code

If the SD card is removed and reinserted please issue the deej.core.reboot command and reopen the serial port
#### Commands
##### deej.core.start
starts up a constant stream of data from the arduino to the host pc. The values are sent as 'x|x|...|x' with x being the analog value
##### deej.core.stop
Stops tje constant steam of slider date from the arduino
##### deej.core.values
Sends one set of values sent as 'x|x|...|x' with x being the analog value
##### deej.core.values.hr
Sends one set of values as 'Slider #n:x mv | Slider #n:x mv | ... | Slider #n:x mv' with n being the slider number and x being the value
##### deej.core.reboot
Reboots the microcontroler (the serial port will have to be reopened)
##### deej.modules.TSC9548A.select
Selects a port on the TSC9548A port range is 0-7 and send the port number as a new line
##### deej.modules.display.setimage
Sets a image on the display. Following this command send the filename on a new line
##### deej.modules.display.off
Turn a display off. The image is keept in the displays ram so no need to resend the image
##### deej.modules.display.on
Turn a display on. The image is keept in the displays ram so no need to resend the image
##### deej.modules.sd.send
Send a file over command line to the sd card. Following this command send the file name on a new line. Then send the bytes raw followed by EOF as chars. Your file cannot contain EOF next to each other but this is unlikely if it isnt a text file
##### deej.modules.sd.list
List the files on the sd card
##### deej.modules.sd.delete
delete a file on the sd card. Following this command send the file name on a new line