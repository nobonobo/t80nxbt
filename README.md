# t80nxbt

Thurastmaster-T80 to ProCon Converter.

## connection

T80----usb----RaspberryPi4-----bluetooth-----NintendoSwitch

## setup RasPi OS

- use RapberryPi Imager, write RpiOS Lite(32bits)
- config hostname and Wi-Fi, ssh
- sudo apt update
- sudo apt install -y python3-pip
- sudo pip3 install nxbt

## build and install

```
make generate build
scp {t80nxbt,procon.py} pi@hostname.local:/home/pi/
```

/etc/systemd/system/t80nxbt.service

```
[Unit]
Description=Rpi ProCon Emulator
After=bluetooth.target

[Service]
ExecStart=/home/pi/t80nxbt -script=/home/pi/procon.py
Restart=always

[Install]
WantedBy=multi-user.target
```

```
sudo systemctl daemon-reload
sudo systemctl enable rpicon
sudo systemctl start rpicon
```

## using

- Nintendo Switch standby for connect new controller.
- connect usb game controller.
- wait few seconds, you can see controller.
- you can use game controller as procon!

## mapping

- L2: fullbrake
- R2: fullaccel
- LeftPaddle: L-button
- RightPaddle: R-button
- ○: A-button
- ×: B-button
- △: X-button
- □: Y-button
- L3: ZL-button
- R3: ZR-button
- SHARE: Minus-button
- OPTIONS: Plus-button
- PS+SHARE: Capture-button
- PS+OPTIONS: Home-button
- PS+L3: LStick-button
- PS+R3: RStick-button
- Wheel: LStick-X(range:-100..100)
- LeftPeddale: RStick-X(range:0..100)
- RightPeddale: RStick-Y(range:0..100)
