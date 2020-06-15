# deejdsp
## This is a deej project that uses custom written plugins to support displays!
This uses my modifed fork of deej. This is still in the development phase but expect to see a more finished version soon!
## TODO
1. Add Taskbar send File
2. Get a image from a session
3. Arrange a image in the center of the screen
4. Convert image to .b file format[See jax-b\ssd1306FilePrep] (https://github.com/jax-b/ssd1306FilePrep]
5. Add config file option for auto generate image (req 1-3)
  - this will use a sha-1 hash of the process name truncated to 8 characters in order to prevent confilt with user generated images
  - if the file does not exist on the end device it will be generated and sent
6. Find free images for other deej config options (master, system, mic) 
## You can view my wireing guide
![EasyEDA](https://image.easyeda.com/histories/df4c1db5c05449faacae832d4a9c00cf.png)
You will also need a sd card adapter in order to store your images. This can be scaled from two to eight of sliders and displays. With some work it can also be scalled far beyond. 

My whole PCB and wireing project can be found on [EasyEDA](https://easyeda.com/jackson_6/deej5a)

#### Original code is by [Omri Harel](deej.rocks)
#### The Main Development also has a [discord](https://discord.gg/nf88NJu)
[![Discord](https://img.shields.io/discord/702940502038937667?logo=discord)](https://discord.gg/nf88NJu)

[POC Video](assets/POC.mkv)
