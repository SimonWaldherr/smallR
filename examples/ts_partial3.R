ma <- function(x, n) {
  result <- c()
  L <- length(x)
  for (i in n:L) {
    w <- x[(i-n+1):i]
    result <- c(result, mean(w))
  }
  result
}
