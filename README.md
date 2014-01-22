WowSuchTicker is a simple program which calls on the public Cryptsy API and gets the current
DOGE/BTC exchange rate.

It was written to be called periodically from conky.   It contains some structs to Marshall the
data from Cryptsy - if some work was done to convert the strings (the API returns nothing but
string values) into the appropriate types, they could be useful for people who want to use
golang to deal with the API.

There are no external dependencies, everything called by WowSuchTicker is from the standard libs.
