# Functions examples

square <- function(x) {
  x * x
}
print(square(3))

# default parameter expression using another param
inc <- function(x, by = x + 1) {
  x + by
}
print(inc(4))
print(inc(4, 10))

# anonymous function and immediate call
print((function(a, b) { a * b })(6, 7))
