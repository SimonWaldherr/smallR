
# smallR demo

x <- c(1, 2, 3, NA, 5)
print(x)
print(is.na(x))
print(sum(x))
print(sum(x, na.rm=TRUE))
print(mean(x, na.rm=TRUE))

y <- 1:5
print(y[2])
print(y[c(1, 3, 5)])

f <- function(a, b = a + 1) {
  a + b
}
print(f(10))
print(f(10, 100))

lst <- list(a = 1, b = "hi")
print(lst$a)
lst$c <- 42
print(names(lst))
print(lst[[3]])


df <- data.frame(a = 1:5, b = c(10, 20, 30, 40, 50))
print(class(df))
print(nrow(df))
print(ncol(df))
print(dim(df))
print(head(df, 2)$a)
print(tail(df, 2)$b)
