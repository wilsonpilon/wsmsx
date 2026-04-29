100 FOR I = 0 TO 4: STRIG(I) ON: NEXT
110 ON STRIG GOSUB 180, 190, 200, 210, 220
120 CLS
130 PRINT"Press spacebar or a button joystick"
140 PRINT
150 GOTO 170
160 PRINT " pressed"
170 GOTO 170
180 PRINT "Spacebar";:RETURN 160
190 PRINT "Button 1 of joystick 1";:RETURN 160
200 PRINT "Button 1 of joystick 2";:RETURN 160
210 PRINT "Button 2 of joystick 1";:RETURN 160
220 PRINT "Button 2 of joystick 2";:RETURN 160

