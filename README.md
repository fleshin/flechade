# flechade
Pack and distribute customizations and themes. Consitent themes all across your Debian box: gnome, gdm, grub, plymouth,etc. Put all your resources on an git repo, define a simple yaml file and apply the customizations consistently on multiple boxes. 

[![Go Report Card](https://goreportcard.com/badge/github.com/fleshin/flechade)](https://goreportcard.com/report/github.com/fleshin/flechade)

## Motivation
Everytime I distro hop, get a new computer or install somebody else's computer, I need to spend a good amount of time customizing it to feel at home.
This is not only time consuming but also incosistent as I forget steps or make mistakes. I wanted a way to keep a consistent enviroment across time and boxes.
I think creating a distro for this purpose is an overkill, as I am not interested on messing up with packaging, booting, etc.
I know most ditros are just that, a tuned version of a "core" distro, so I would try to avoid them. Then we have three main camps:

- Fedora / RH (rpm)
- Arch
- Debian

IMHO, RedHat is has a record of messing up things and I think .deb is far more popular than rpm. Arch is ideal for customization, but might be to much of a moving target to build on top of. So, Debian stable is great for stability but might get on the oudated side, while debian testing would be my sweet spot regarding stability/freshness (feel free to disagree on that, this is just my personal preference)

## Solution 
The tool reads a GIT repository with a single YAML file and resources that will coordinate themes and customizations across the system such as grub2 backgrounds, plymouth themes, gdm3/lightdm, gnome extenstions, polybar themes, custom shells, extra fonts, etc. 
There is no need for scripting skills and anyone can fork one of the examples and modify at will from there.
The YAML file contains simple "commands" that allow to download files, copy, etc while hidding all the details and complexity (check the flechade.yaml file from the example repos).

## Build
Build and install through Go toolset:
```
sudo apt install golang
go install github.com/fleshin/flechade@latest
```

Or, clone the repo and build:
```
git  clone http://github.com/fleshin/flechade && cd flechade
go build
```

## Run
On a freshly installed Debian 12 or testing, open a terminal and run below command. Take into account that this process will overwrite apt sources, configuration files and other resources.

```
su - root -c "usermod -aG sudo $USER"; newgrp sudo; 
sudo ~/go/bin/flechade -l
```

## Usage
Load default customization set (Golang McGamer)
```
sudo ./flechade -l
```
Load customization set from a directory
```
sudo ./flechade -d /tmp/custom-flechade
```
Load customization from a GIT repository
```
sudo ~/go/bin/flechade -r https://github.com/fleshin/flechade-normie
```

## Sreenshots of Golang MacGamer (default)

<p align="center"> <img src="https://raw.githubusercontent.com/fleshin/fleshin/master/ss2.png"/> </p>

<p align="center"> <img src="https://raw.githubusercontent.com/fleshin/fleshin/master/ss1.png"/> </p>

<p align="center"> <img src="https://raw.githubusercontent.com/fleshin/fleshin/master/ss5.png"/> </p>

## Sreenshots of [Tracie Progsock](https://github.com/fleshin/flechade-tracie)

<p align="center"> <img src="https://raw.githubusercontent.com/fleshin/fleshin/master/tp1.png"/> </p>

<p align="center"> <img src="https://raw.githubusercontent.com/fleshin/fleshin/master/tp2.png"/> </p>

<p align="center"> <img src="https://raw.githubusercontent.com/fleshin/fleshin/master/tp3.png"/> </p>

## Sreenshots of [Normie Winoffice](https://github.com/fleshin/flechade-normie)

<p align="center"> <img src="https://raw.githubusercontent.com/fleshin/fleshin/master/nw1.png"/> </p>

<p align="center"> <img src="https://raw.githubusercontent.com/fleshin/fleshin/master/nw2.png"/> </p>

<p align="center"> <img src="https://raw.githubusercontent.com/fleshin/fleshin/master/nw3.png"/> </p>

<p align="center"> <img src="https://raw.githubusercontent.com/fleshin/fleshin/master/nw4.png"/> </p>
