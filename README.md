# musiccast-linker
simple cli to enable yamaha musiccast link feature without touching your phone.
I use it to link a pair of mc-20 to the digital input of the second zone of my yamaha rx reciever.

## flags 
````
  -clients string
        comma separated list of client hostnames
  -master string
        master hostname
  -master-input string
        (optional) set streaming input for given zone
  -master-zone string
        master zone to link (default "zone2")
  -standby
        set this to power off clients and master

````
## example
#### setup link
```
musiccast-link -master rx-a780 -master-zone zone2 -master-input audio2 -clients mc-20-kitchen,mc20-office
```
#### standby
```
musiccast-link -standby -master rx-a780 -master-zone zone2 -clients mc-20-kitchen,mc20-office
```