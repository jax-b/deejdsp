#include "arduino.h"
#include <SPI.h>
#include <Wire.h>
#include <SD.h>
#include "ssd1306CMDS.h"
#include <avr/wdt.h>

//You must Hard Code in the number of Sliders in
#define NUM_SLIDERS 6
#define SERIALSPEED 115200
#define NUM_DISPLAYS 6
#define SerialTimeout 2000

const uint8_t analogInputs[NUM_SLIDERS] = {18, 19, 20, 21, 9, 8};

//IIC and I2C are the same thing
#define IICMULTIPLEXADDR 0x70
#define IICDSPADDR 0x3C

// GFX Settings
#define SCREEN_WIDTH 128 // OLED display width, in pixels
#define SCREEN_HEIGHT 64 // OLED display height, in pixels

#define SDCSPIN 7

uint16_t analogSliderValues[NUM_SLIDERS];

// Constend Send
bool pushSliderValuesToPC = false;

void setup() { 
  for (uint8_t i = 0; i < NUM_SLIDERS; i++) {
    pinMode(analogInputs[i], INPUT);
  }
  Wire.begin();
  Serial.begin(SERIALSPEED);
  Serial.setTimeout(2000);
  Serial.print("INITBEGIN");
  for (int i = 0; i < NUM_DISPLAYS; i++) {
    Serial.print("DSP" + String(i) + "INIT ");
    tcaselect(IICMULTIPLEXADDR, i);
    dspInit(IICDSPADDR);
    dspClear(IICDSPADDR);
//    dspSendData(IICDSPADDR, i+1);
  }
  Serial.print("SDINIT ");
  if (!SD.begin(SDCSPIN)){
    Serial.println("SDERROR ");
    while(1);
  }
  Serial.println("INITDONE");
}

void loop() {
  checkForCommand();

  updateSliderValues();

  //Check for data chanel to be open
  if(pushSliderValuesToPC) {
    sendSliderValues(); // Actually send data
  }
  // printSliderValues(); // For debug
  delay(10);
}

void reboot() {
  wdt_disable();
  wdt_enable(WDTO_30MS);
  while (1) {}
}

void updateSliderValues() {
  for (uint8_t i = 0; i < NUM_SLIDERS; i++) {
     analogSliderValues[i] = analogRead(analogInputs[i]);
  }
}

void sendSliderValues() {
  String builtString = String("");

  for (uint8_t i = 0; i < NUM_SLIDERS; i++) {
    builtString += String((int)analogSliderValues[i]);

    if (i < NUM_SLIDERS - 1) {
      builtString += String("|");
    }
  }
  
  Serial.println(builtString);
}

void printSliderValues() {
  for (uint8_t i = 0; i < NUM_SLIDERS; i++) {
    Serial.print("Slider #"+ String(i + 1) + ": " + String(analogSliderValues[i]) + " mV");

    if (i < NUM_SLIDERS - 1) {
      Serial.print(" | ");
    } else {
      Serial.println();
    }
  }
}

void checkForCommand() {
  //Check if data is waiting
  if (Serial.available() > 0) {
    //Get start time of command
    unsigned long timeStart = millis();

    //Get data from Serial
    String input = Serial.readStringUntil('\n');  // Read chars from serial monitor

    //If data takes to long
    if(millis()-timeStart >= SerialTimeout) {
      Serial.println("TIMEOUT");
      return;
    }

    // Check and match commands
    else {

      // Start Sending Slider Values
      if ( input.equalsIgnoreCase("deej.core.start") == true ) {
        pushSliderValuesToPC = true;
      }

      // Stop Sending Slider Values
      else if ( input.equalsIgnoreCase("deej.core.stop") == true ) {
        pushSliderValuesToPC = false;
      }
      
      // Send Single Slider Values
      else if ( input.equalsIgnoreCase("deej.core.values") == true ) {
        sendSliderValues();
      }

      // Send Human Readable Slider Values 
      else if ( input.equalsIgnoreCase("deej.core.values.HR") == true ) {
        printSliderValues();
      }

      // Reboot MCU
      else if ( input.equalsIgnoreCase("deej.core.reboot") == true ) {
        reboot();
      }

      // Select Port on TSC9548A
      // Following this command send the port number on a new line
      else if ( input.equalsIgnoreCase("deej.modules.TSC9548A.select")== true) {
        timeStart = millis();

        //Get data from Serial
        uint8_t portnumber = Serial.readStringUntil('\n').toInt();  // Read chars from Serial monitor
        
        //Get Stop Time and exit if timeout
        if(millis()-timeStart >= SerialTimeout) {
          Serial.println("TIMEOUT");
        }
        else {
          tcaselect(IICMULTIPLEXADDR, portnumber);
        }
      }

      // Set image on display from file
      else if ( input.equalsIgnoreCase("deej.modules.display.setimage") == true ){
        timeStart = millis();

        //Get data from Serial
        String filename = Serial.readStringUntil('\n');  // Read chars from Serial monitor
        
        if(millis()-timeStart >= SerialTimeout) {
          Serial.println("TIMEOUT");
        }
        else {
          if (!SD.exists(filename)){
            Serial.print("FILENOTFOUND");
          }
          else {
            dspSetImage(IICDSPADDR,filename);
          }
        }
      }

      // Turn a display off
      // Image is keept in the displays ram so no need to resend the image
      else if ( input.equalsIgnoreCase("deej.modules.display.off") == true) {
        dspOff(IICDSPADDR);
      }

      // Turn a display on
      // Image is keept in the displays ram so no need to resend the image
      else if ( input.equalsIgnoreCase("deej.modules.display.on") == true) {
        dspOn(IICDSPADDR);
      }

      // Send a file over command line to the sd card
      // Following this command send the file name on a new line
      // Then send the bytes raw followed by EOF as chars
      // your file cannot contain EOF next to eachother but this is unlikely if it isnt a text file
      else if ( input.equalsIgnoreCase("deej.modules.sd.send") == true ){
        timeStart = millis();

        //Get data from Serial
        String filename = Serial.readStringUntil('\n');  // Read chars from Serial monitor
        
        if(timeStart-millis() >= 1000) {
          Serial.println("TIMEOUT");
        }
        else {
          sdPutFile(filename);
        }
      }
      
      // List the files on the sd card
      else if ( input.equalsIgnoreCase("deej.modules.sd.list") == true){
        File root = SD.open("/");
        sdPrintDirectory(root, 0);
      }

      // delete a file on the sd card
      // Following this command send the file name on a new line
      else if ( input.equalsIgnoreCase("deej.modules.sd.delete") == true){
        timeStart = millis();

        //Get data from Serial
        String filename = Serial.readStringUntil('\n');  // Read chars from Serial monitor
        
        if(timeStart-millis() >= 1000) {
          Serial.println("TIMEOUT");
        }
        else {
          sdDelete(filename);
        }
      }
      
      //Default Catch all
      else {
        Serial.println("INVALIDCOMMAND");
      }
      
    }
  }
}

// SD Card List Files
void sdPrintDirectory(File dir, int numTabs) {
  while (true) {

    File entry =  dir.openNextFile();
    if (! entry) {
      // no more files
      break;
    }
    for (uint8_t i = 0; i < numTabs; i++) {
      Serial.print('\t');
    }
    Serial.print(entry.name());
    if (entry.isDirectory()) {
      Serial.println("/");
      sdPrintDirectory(entry, numTabs + 1);
    } else {
      // files have sizes, directories do not
      Serial.print("\t\t");
      Serial.println(entry.size(), DEC);
    }
    entry.close();
  }
  Serial.println("DONE");
}

// SD Card send file
void sdPutFile(const String filename) {
  Serial.println("Starting File Write");
  Serial.println("Waiting for EOF");
  File imgFile = SD.open(filename, FILE_WRITE);
  int16_t last3[3] = {-1,-1,-1};
  while ( last3[0] != 'E' && last3[1] != 'O' && last3[2] != 'F' ) {
    if ( last3[0] != -1 ) {
      imgFile.write(last3[0]);
    }
    last3[0] = last3[1];
    last3[1] = last3[2];
    int nextByte = Serial.read();
    if (nextByte != -1) {
      last3[2] = nextByte;
    }
  }
  while(Serial.available() > 0) {
    Serial.read();
  }
  imgFile.close();
  Serial.print("EOFDETECT");
}

// SD Card delete file
void sdDelete(const String filename) {
  if (SD.exists(filename)){
    Serial.print("FILENOTFOUND");
  }
  else {
    SD.remove(filename);
    Serial.print("FILEDELETED");
  }
}

// TSC9548A IIC,IIC,I2C multiplexor port select 
void tcaselect(uint8_t addr, uint8_t i) {
  if (i > 7) return;
 
  Wire.beginTransmission(addr);
  Wire.write(1 << i);
  Wire.endTransmission();  
}

// Display Module Start
// Send a Command to the ssd1306 
void dspSendCommand(uint8_t addr, uint8_t c){
  Wire.beginTransmission(addr);
  Wire.write(0x00);
  Wire.write(c);
  Wire.endTransmission();
}

// Send Display Data to ssd1306
void dspSendData(uint8_t addr, uint8_t c){
  Wire.beginTransmission(addr);
  Wire.write(0x40);
  Wire.write(c);
  Wire.endTransmission();
}

// initialize SSD1306 display at an address
void dspInit(uint8_t addr){
  // ssd1306 Display initialization sequence
  // see this page for the sequence for the sequence i used:
  // https://iotexpert.com/2019/08/07/debugging-ssd1306-display-problems/

  const char initializeCmds[]={
    //////// Fundamental Commands
    OLED_DISPLAYOFF,          // 0xAE Screen Off
    OLED_SETCONTRAST,         // 0x81 Set contrast control
    0x7F,                     // 0-FF ... default half way
    OLED_DISPLAYNORMAL,       // 0xA6, //Set normal display 
    //////// Scrolling Commands
    OLED_DEACTIVATE_SCROLL,   // Deactive scroll
    //////// Addressing Commands
    OLED_SETMEMORYMODE,       // 0x20, //Set memory address mode
    OLED_SETMEMORYMODE_HORIZONTAL,  // Page
    //////// Hardware Configuration Commands
    OLED_SEGREMAPINV,         // 0xA1, //Set segment re-map 
    OLED_SETMULTIPLEX,        // 0xA8 Set multiplex ratio
    0x3F,                     // Vertical Size - 1
    OLED_COMSCANDEC,          // 0xC0 Set COM output scan direction
    OLED_SETDISPLAYOFFSET,    // 0xD3 Set Display Offset
    0x00,                     //
    OLED_SETCOMPINS,          // 0xDA Set COM pins hardware configuration
    0x12,                     // Alternate com config & disable com left/right
    //////// Timing and Driving Settings
    OLED_SETDISPLAYCLOCKDIV,  // 0xD5 Set display oscillator frequency 0-0xF /clock divide ratio 0-0xF
    0x80,                     // Default value
    OLED_SETPRECHARGE,        // 0xD9 Set pre-changed period
    0x22,                     // Default 0x22
    OLED_SETVCOMDESELECT,     // 0xDB, //Set VCOMH Deselected level
    0x20,                     // Default 
    //////// Charge pump regulator
    OLED_CHARGEPUMP,          // 0x8D Set charge pump
    OLED_CHARGEPUMP_ON,       // 0x14 VCC generated by internal DC/DC circuit
    // Turn the screen back on...       
    OLED_DISPLAYALLONRESUME,  // 0xA4, //Set entire display on/off
    OLED_DISPLAYON,           // 0xAF  //Set display on
  };
  
  for(char i=0;i<25;i++){
    dspSendCommand(addr, initializeCmds[i]);
  }
}

// set the column 
// ref the ssd 1306 datasheet if you want to find out how it works
void dspSetColumn(uint8_t addr, uint8_t cstart, uint8_t cend) {
  dspSendCommand(addr, 0x21);
  dspSendCommand(addr, cstart);
  dspSendCommand(addr, cend);
}

// set the page
// ref the ssd 1306 datasheet if you want to find out how it works
void dspSetPage(uint8_t addr, uint8_t pstart, uint8_t pend) {
  dspSendCommand(addr, 0x22);
  dspSendCommand(addr, pstart);
  dspSendCommand(addr, pend);
}

void dspOff(uint8_t addr){
  dspSendCommand(addr, OLED_DISPLAYOFF);
}

// turns the specified display on
void dspOn(uint8_t addr){
  dspSendCommand(addr, OLED_DISPLAYON);
}

void dspClear(uint8_t addr){
  // go to zero and set end to full end
  dspSetColumn(addr, 0x00,0x7F);
  // go to zero and set end to full end
  dspSetPage(addr, 0xB0,0xB7);
  // fill the GFX Ram on the ssd1306 with zeros blanking the display
  for(int i = 0;i < (SCREEN_WIDTH * (SCREEN_HEIGHT/8)); i++){
    dspSendData(addr, 0b00000000);
  }
}

// Writes a image to the ssd1306 display
void dspSetImage(uint8_t addr, String imagefilename) {
  // open the image file
  // also this file should almost allways contain 8192 bytes
  File imgFile = SD.open(imagefilename);
  // clear the display
  dspClear(addr);
  // initialize some temp vars
  int inputChar;
  int maxPages = 8;

  // loop through each page 
  // each padge is 8 Vertical bytes per column
  // we write to 128 columns each column is 8 bytes tall or 8 pixel.
  // there are 8 pages [0-7] to make up the 64 pixel tall display
  // we also process all posable ascii char including newline and carrage return
  // since a char is one byte it makes it easy to read data from the file and into the buffer
  while (maxPages != 0 && inputChar != -1){
    int CharsLeftInLine = 128;
    while  (CharsLeftInLine > 0 && inputChar != -1){
      inputChar = imgFile.read();
      Serial.print(char(inputChar));
      if(inputChar == -1){
        break;
      }
      dspSendData(addr, inputChar);
      CharsLeftInLine--;
    }
    Serial.println();
    maxPages--;
  }
  imgFile.close();
  Serial.println("\nDSPWRITEDONE");
}