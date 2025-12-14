# Lists and vectors
v <- c(10, 20, NA, 40)
print(v)
print(is.na(v))
print(sum(v, na.rm=TRUE))

lst <- list(name = "alice", score = 99)
print(lst$name)
lst$extra <- "note"
print(names(lst))
print(lst[[3]])
