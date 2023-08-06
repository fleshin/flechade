# flechade
Customize your linux box

## Motivation
Everytime I distro hop, get a new computer or install somebody else's computer, I need to spend a good amount of time customizing it to feel at home.
This is not only time consuming but also incosistent as I forget steps or make mistakes. I wanted a way to keep a consistent enviroment across time and boxes.
I think creating a distro for this purpose is an overkill, as I am not interested on messing up with packaging, booting, etc.
I know most ditros are just that, a tuned version of a "core" distro, so I would try to avoid them. Then we have three main camps:

- Fedora / RH (rpm)
- Arch
- Debian

IMHO, RedHat is has a record of messing up things and I think .deb is far more popular than rpm. Arch is ideal for customization, but might be to much of a moving target to build on top of. So, Debian stable is great for stability but might get on the oudated side, while debian testing would be my sweet spot regarding stability/freshness (feel free to disagree on that, this is just my personal preference)

## Build
```
go install github.com/fleshin/flechade@latest
```

## Run
On a freshly installed debian 12 or testing, open a terminal and run this command:

```
su - root -c "usermod -aG sudo $USER"; newgrp sudo; 
sudo flechade
```

## TODO
Implement a single file customization package (yaml?) to be processed by the executable

## Examples of default customization


<p align="center"> <img src="https://raw.githubusercontent.com/fleshin/fleshin/master/ss1.png"/> </p>

<p align="center"> <img src="https://raw.githubusercontent.com/fleshin/fleshin/master/ss2.png"/> </p>

<p align="center"> <img src="https://raw.githubusercontent.com/fleshin/fleshin/master/ss3.png"/> </p>

<p align="center"> <img src="https://raw.githubusercontent.com/fleshin/fleshin/master/ss4.png"/> </p>

<p align="center"> <img src="https://raw.githubusercontent.com/fleshin/fleshin/master/ss5.png"/> </p>


