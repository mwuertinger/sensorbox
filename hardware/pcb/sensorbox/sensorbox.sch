EESchema Schematic File Version 4
EELAYER 30 0
EELAYER END
$Descr A4 11693 8268
encoding utf-8
Sheet 1 1
Title ""
Date ""
Rev ""
Comp ""
Comment1 ""
Comment2 ""
Comment3 ""
Comment4 ""
$EndDescr
$Comp
L sensorbox:NodeMCU_D1 U1
U 1 1 5F86460D
P 5150 2000
F 0 "U1" H 5578 2046 50  0000 L CNN
F 1 "NodeMCU_D1" H 5578 1955 50  0000 L CNN
F 2 "sensorbox:NodeMCU_D1" H 5150 2000 50  0001 C CNN
F 3 "" H 5150 2000 50  0001 C CNN
	1    5150 2000
	1    0    0    -1  
$EndComp
Text GLabel 4750 1500 1    50   Output ~ 0
3V3
Wire Wire Line
	4750 1550 4750 1500
Text GLabel 5250 2500 3    50   Output ~ 0
SCL
Wire Wire Line
	5250 2450 5250 2500
Text GLabel 5150 2500 3    50   BiDi ~ 0
SDA
Wire Wire Line
	5150 2450 5150 2500
Text GLabel 4950 1500 1    50   Input ~ 0
S1
Text GLabel 4850 1500 1    50   Output ~ 0
S2
Wire Wire Line
	4850 1500 4850 1550
Wire Wire Line
	4950 1500 4950 1550
$Comp
L sensorbox:MH-Z14A U2
U 1 1 5F870D57
P 1900 2250
F 0 "U2" V 1335 2283 50  0000 C CNN
F 1 "MH-Z14A" V 1426 2283 50  0000 C CNN
F 2 "sensorbox:MH-Z14A" H 1900 2250 50  0001 C CNN
F 3 "" H 1900 2250 50  0001 C CNN
	1    1900 2250
	0    1    1    0   
$EndComp
Text GLabel 2250 1950 2    50   BiDi ~ 0
5V
Text GLabel 2250 2050 2    50   BiDi ~ 0
GND
Text GLabel 2250 2200 2    50   Input ~ 0
S2
Text GLabel 2250 2300 2    50   Output ~ 0
S1
Wire Wire Line
	2200 1950 2250 1950
Wire Wire Line
	2250 2050 2200 2050
Wire Wire Line
	2200 2200 2250 2200
Wire Wire Line
	2250 2300 2200 2300
Text GLabel 8300 2300 1    50   Input ~ 0
GND
Text GLabel 8400 2300 1    50   Input ~ 0
3V3
Text GLabel 8500 2300 1    50   Input ~ 0
SCL
Text GLabel 8600 2300 1    50   BiDi ~ 0
SDA
Wire Wire Line
	8300 2350 8300 2300
Wire Wire Line
	8400 2350 8400 2300
Wire Wire Line
	8500 2350 8500 2300
Wire Wire Line
	8600 2300 8600 2350
$Comp
L sensorbox:BME280 U3
U 1 1 5F878379
P 6950 2750
F 0 "U3" H 7278 2854 50  0000 L CNN
F 1 "BME280" H 7278 2763 50  0000 L CNN
F 2 "sensorbox:BME280" H 7000 2750 50  0001 C CNN
F 3 "" H 7000 2750 50  0001 C CNN
	1    6950 2750
	1    0    0    -1  
$EndComp
$Comp
L sensorbox:OLED-Display-0.69 U4
U 1 1 5F875EE2
P 8450 2800
F 0 "U4" H 9278 2854 50  0000 L CNN
F 1 "OLED-Display-0.69" H 9278 2763 50  0000 L CNN
F 2 "sensorbox:OLED-Display-0.69" H 8450 2800 50  0001 C CNN
F 3 "" H 8450 2800 50  0001 C CNN
	1    8450 2800
	1    0    0    -1  
$EndComp
Text GLabel 6800 2400 1    50   Input ~ 0
3V3
Text GLabel 6900 2400 1    50   Input ~ 0
GND
Text GLabel 7000 2400 1    50   Input ~ 0
SCL
Text GLabel 7100 2400 1    50   BiDi ~ 0
SDA
Wire Wire Line
	7100 2450 7100 2400
Wire Wire Line
	7000 2450 7000 2400
Wire Wire Line
	6900 2450 6900 2400
Wire Wire Line
	6800 2450 6800 2400
$EndSCHEMATC
