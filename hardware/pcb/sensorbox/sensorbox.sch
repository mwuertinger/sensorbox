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
P 6800 2200
F 0 "U3" H 7128 2304 50  0000 L CNN
F 1 "BME280" H 7128 2213 50  0000 L CNN
F 2 "sensorbox:BME280" H 6850 2200 50  0001 C CNN
F 3 "" H 6850 2200 50  0001 C CNN
	1    6800 2200
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
Text GLabel 6650 1850 1    50   Input ~ 0
3V3
Text GLabel 6750 1850 1    50   Input ~ 0
GND
Text GLabel 6850 1850 1    50   Input ~ 0
SCL
Text GLabel 6950 1850 1    50   BiDi ~ 0
SDA
Wire Wire Line
	6950 1900 6950 1850
Wire Wire Line
	6850 1900 6850 1850
Wire Wire Line
	6750 1900 6750 1850
Wire Wire Line
	6650 1900 6650 1850
$Comp
L Switch:SW_Push SW1
U 1 1 5F87CC89
P 3450 3150
F 0 "SW1" H 3450 3435 50  0000 C CNN
F 1 "Display" H 3450 3344 50  0000 C CNN
F 2 "Button_Switch_THT:SW_PUSH_6mm" H 3450 3350 50  0001 C CNN
F 3 "~" H 3450 3350 50  0001 C CNN
	1    3450 3150
	1    0    0    -1  
$EndComp
Text GLabel 4850 2500 3    50   Output ~ 0
GND
Wire Wire Line
	4850 2450 4850 2500
Text GLabel 3100 3150 0    50   Output ~ 0
GND
Wire Wire Line
	3100 3150 3250 3150
Text GLabel 3900 3150 2    50   Output ~ 0
DISPLAY
Text GLabel 5050 2500 3    50   Input ~ 0
DISPLAY
Wire Wire Line
	5050 2450 5050 2500
Text GLabel 4750 2500 3    50   Output ~ 0
5V
Wire Wire Line
	4750 2450 4750 2500
$Comp
L Device:LED D3
U 1 1 5F884E0D
P 4700 4600
F 0 "D3" V 4739 4483 50  0000 R CNN
F 1 "LED" V 4648 4483 50  0000 R CNN
F 2 "LED_THT:LED_D3.0mm" H 4700 4600 50  0001 C CNN
F 3 "~" H 4700 4600 50  0001 C CNN
	1    4700 4600
	0    -1   -1   0   
$EndComp
Wire Wire Line
	4700 4750 4700 4850
$Comp
L Device:R R5
U 1 1 5F88657F
P 4700 4250
F 0 "R5" H 4770 4296 50  0000 L CNN
F 1 "330" H 4770 4205 50  0000 L CNN
F 2 "Resistor_SMD:R_0805_2012Metric" V 4630 4250 50  0001 C CNN
F 3 "~" H 4700 4250 50  0001 C CNN
	1    4700 4250
	1    0    0    -1  
$EndComp
Wire Wire Line
	4700 4400 4700 4450
$Comp
L Device:R R1
U 1 1 5F841061
P 3150 1600
F 0 "R1" H 3220 1646 50  0000 L CNN
F 1 "4.7k" H 3220 1555 50  0000 L CNN
F 2 "Resistor_SMD:R_0805_2012Metric" V 3080 1600 50  0001 C CNN
F 3 "~" H 3150 1600 50  0001 C CNN
	1    3150 1600
	1    0    0    -1  
$EndComp
$Comp
L Device:R R2
U 1 1 5F84144D
P 3450 1600
F 0 "R2" H 3520 1646 50  0000 L CNN
F 1 "4.7k" H 3520 1555 50  0000 L CNN
F 2 "Resistor_SMD:R_0805_2012Metric" V 3380 1600 50  0001 C CNN
F 3 "~" H 3450 1600 50  0001 C CNN
	1    3450 1600
	1    0    0    -1  
$EndComp
Text GLabel 3000 1300 0    50   Input ~ 0
3V3
Wire Wire Line
	3000 1300 3150 1300
Wire Wire Line
	3450 1300 3450 1450
Wire Wire Line
	3150 1450 3150 1300
Connection ~ 3150 1300
Wire Wire Line
	3150 1300 3450 1300
Text GLabel 3150 1800 3    50   Input ~ 0
SDA
Text GLabel 3450 1800 3    50   Input ~ 0
SCL
Wire Wire Line
	3150 1750 3150 1800
Wire Wire Line
	3450 1750 3450 1800
Wire Wire Line
	3650 3150 3900 3150
$Comp
L Transistor_FET:2N7002 Q3
U 1 1 5F8C779D
P 4600 5050
F 0 "Q3" H 4804 5096 50  0000 L CNN
F 1 "2N7002" H 4804 5005 50  0000 L CNN
F 2 "Package_TO_SOT_SMD:SOT-23" H 4800 4975 50  0001 L CIN
F 3 "https://www.fairchildsemi.com/datasheets/2N/2N7002.pdf" H 4600 5050 50  0001 L CNN
	1    4600 5050
	1    0    0    -1  
$EndComp
Text GLabel 4700 4050 1    50   Input ~ 0
5V
Wire Wire Line
	4700 4050 4700 4100
Text GLabel 4700 5300 3    50   Input ~ 0
GND
Wire Wire Line
	4700 5250 4700 5300
$Comp
L Device:LED D1
U 1 1 5F8EE25E
P 7000 4600
F 0 "D1" V 7039 4483 50  0000 R CNN
F 1 "LED" V 6948 4483 50  0000 R CNN
F 2 "LED_THT:LED_D3.0mm" H 7000 4600 50  0001 C CNN
F 3 "~" H 7000 4600 50  0001 C CNN
	1    7000 4600
	0    -1   -1   0   
$EndComp
Wire Wire Line
	7000 4750 7000 4850
$Comp
L Device:R R3
U 1 1 5F8EE265
P 7000 4250
F 0 "R3" H 7070 4296 50  0000 L CNN
F 1 "330" H 7070 4205 50  0000 L CNN
F 2 "Resistor_SMD:R_0805_2012Metric" V 6930 4250 50  0001 C CNN
F 3 "~" H 7000 4250 50  0001 C CNN
	1    7000 4250
	1    0    0    -1  
$EndComp
Wire Wire Line
	7000 4400 7000 4450
$Comp
L Transistor_FET:2N7002 Q1
U 1 1 5F8EE26C
P 6900 5050
F 0 "Q1" H 7104 5096 50  0000 L CNN
F 1 "2N7002" H 7104 5005 50  0000 L CNN
F 2 "Package_TO_SOT_SMD:SOT-23" H 7100 4975 50  0001 L CIN
F 3 "https://www.fairchildsemi.com/datasheets/2N/2N7002.pdf" H 6900 5050 50  0001 L CNN
	1    6900 5050
	1    0    0    -1  
$EndComp
Text GLabel 7000 4050 1    50   Input ~ 0
5V
Wire Wire Line
	7000 4050 7000 4100
Text GLabel 7000 5300 3    50   Input ~ 0
GND
Wire Wire Line
	7000 5250 7000 5300
$Comp
L Device:LED D2
U 1 1 5F8F0CFC
P 5850 4600
F 0 "D2" V 5889 4483 50  0000 R CNN
F 1 "LED" V 5798 4483 50  0000 R CNN
F 2 "LED_THT:LED_D3.0mm" H 5850 4600 50  0001 C CNN
F 3 "~" H 5850 4600 50  0001 C CNN
	1    5850 4600
	0    -1   -1   0   
$EndComp
Wire Wire Line
	5850 4750 5850 4850
$Comp
L Device:R R4
U 1 1 5F8F0D03
P 5850 4250
F 0 "R4" H 5920 4296 50  0000 L CNN
F 1 "330" H 5920 4205 50  0000 L CNN
F 2 "Resistor_SMD:R_0805_2012Metric" V 5780 4250 50  0001 C CNN
F 3 "~" H 5850 4250 50  0001 C CNN
	1    5850 4250
	1    0    0    -1  
$EndComp
Wire Wire Line
	5850 4400 5850 4450
$Comp
L Transistor_FET:2N7002 Q2
U 1 1 5F8F0D0A
P 5750 5050
F 0 "Q2" H 5954 5096 50  0000 L CNN
F 1 "2N7002" H 5954 5005 50  0000 L CNN
F 2 "Package_TO_SOT_SMD:SOT-23" H 5950 4975 50  0001 L CIN
F 3 "https://www.fairchildsemi.com/datasheets/2N/2N7002.pdf" H 5750 5050 50  0001 L CNN
	1    5750 5050
	1    0    0    -1  
$EndComp
Text GLabel 5850 4050 1    50   Input ~ 0
5V
Wire Wire Line
	5850 4050 5850 4100
Text GLabel 5850 5300 3    50   Input ~ 0
GND
Wire Wire Line
	5850 5250 5850 5300
Text GLabel 4350 5050 0    50   Input ~ 0
RED
Wire Wire Line
	4350 5050 4400 5050
Text GLabel 6650 5050 0    50   Input ~ 0
GREEN
Wire Wire Line
	6650 5050 6700 5050
Text GLabel 5500 5050 0    50   Input ~ 0
YELLOW
Wire Wire Line
	5500 5050 5550 5050
Text GLabel 5050 1500 1    50   Output ~ 0
GREEN
Text GLabel 5150 1500 1    50   Output ~ 0
YELLOW
Text GLabel 5250 1500 1    50   Output ~ 0
RED
NoConn ~ 5350 1550
NoConn ~ 5450 1550
NoConn ~ 4950 2450
NoConn ~ 5350 2450
NoConn ~ 5450 2450
NoConn ~ 1550 1900
NoConn ~ 1550 1950
NoConn ~ 1550 2000
NoConn ~ 1550 2050
NoConn ~ 1550 2100
NoConn ~ 1550 2150
NoConn ~ 1550 2200
NoConn ~ 1550 2250
NoConn ~ 1550 2350
NoConn ~ 1550 2400
NoConn ~ 1550 2450
NoConn ~ 1550 2500
Wire Wire Line
	5250 1500 5250 1550
Wire Wire Line
	5150 1500 5150 1550
Wire Wire Line
	5050 1500 5050 1550
Text GLabel 8300 3550 1    50   Input ~ 0
GND
Text GLabel 8400 3550 1    50   Input ~ 0
3V3
Text GLabel 8500 3550 1    50   Input ~ 0
SCL
Text GLabel 8600 3550 1    50   BiDi ~ 0
SDA
Wire Wire Line
	8300 3600 8300 3550
Wire Wire Line
	8400 3600 8400 3550
Wire Wire Line
	8500 3600 8500 3550
Wire Wire Line
	8600 3550 8600 3600
$Comp
L sensorbox:OLED-Display-0.69 U5
U 1 1 5F98F406
P 8450 4050
F 0 "U5" H 9278 4104 50  0000 L CNN
F 1 "OLED-Display-0.69" H 9278 4013 50  0000 L CNN
F 2 "sensorbox:OLED-Display-0.69" H 8450 4050 50  0001 C CNN
F 3 "" H 8450 4050 50  0001 C CNN
	1    8450 4050
	1    0    0    -1  
$EndComp
$EndSCHEMATC
