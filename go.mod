module peterdekok.nl/net

go 1.13

require (
	peterdekok.nl/config v0.0.0
	peterdekok.nl/logger v0.0.0
	peterdekok.nl/trap v0.0.0
)

replace (
	peterdekok.nl/config => ../config
	peterdekok.nl/logger => ../logger
	peterdekok.nl/trap => ../trap
)
