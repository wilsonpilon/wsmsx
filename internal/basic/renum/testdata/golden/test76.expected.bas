100 FOR I=1 TO 3: KEY(I) STOP: NEXT ' F1 to F3 are stopped
110 FOR I=4 TO 10: KEY(I) OFF: NEXT ' F4 to F10 are disabled
120 ON KEY GOSUB 260,270,280 ' Subroutine will only be active from line 130
130 CLS
140 PRINT"TEST 1: No effect"
150 PRINT
160 PRINT"Press a function key or ESC key"
170 A$=INPUT$(1)
180 IF A$=CHR$(27) THEN 190 ELSE 170
190 CLS
200 PRINT"TEST 2"
210 PRINT"Press a function key"
220 KEY(1)ON: KEY(2)ON: KEY(3)ON ' F1 to F3 and the subroutine in line 30 are enabled
230 GOTO 250
240 PRINT:PRINT a$+" pressed"
250 GOTO 250
260 A$="F1": RETURN 240
270 A$="F2": RETURN 240
280 A$="F3": RETURN 240

