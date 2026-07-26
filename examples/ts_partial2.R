ma <- function(x, n) {
  result <- c()
  len <- length(x)
  for (i in seq(n, len)) {
    w <- x[(i-n+1):i]
    result <- c(result, mean(w))
  }
  result
}
