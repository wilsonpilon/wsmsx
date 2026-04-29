100 ' FUNCTION KEYS TEST
110 FOR I = 1 TO 3: KEY(I) STOP: NEXT ' F1 to F3 are stopped
120 FOR I = 4 TO 10: KEY(I) OFF: NEXT ' F4 to F10 are disabled
130 ON KEY GOSUB 260, 270, 280 ' Subroutine will only be active from line 130
140 CLS
150 PRINT"TEST 1"
160 PRINT"Press a function key or space..."
170 FOR I = 1 TO 10 : KEY I,"OFF" : NEXT : KEY ON
180 IF STRIG(0) = 0 THEN 180
190 CLS
200 PRINT"TEST 2"
210 PRINT"Press a function key or space"
220 FOR I = 1 TO 10 : KEY I,"F"+STR$(I) : NEXT
230 KEY(1) ON : KEY(2) ON : KEY(3) ON ' F1 to F3 and the subroutine in line 30 are enabled
240 IF STRIG(0) = 0 THEN 240
250 KEY OFF : END
260 PRINT "F1 pressed" : RETURN
270 PRINT "F2 pressed" : RETURN
280 PRINT "F3 pressed" : RETURN

