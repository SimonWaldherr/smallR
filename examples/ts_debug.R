# Debug time series
# Minimal reproduction of demo code

data <- c(10, 12, 11, 13, 15, 14, 16, 18, 17, 19, 20, 22, 21, 23, 25)
window <- 5

idx <- seq(window, length(data))
moving_avg <- sapply(idx, function(i) mean(data[(i-window+1):i]))

cat("Points:", length(data), "\n")
cat("MA window:", window, "\n")
cat("Mean:", round(mean(data), 2), "\n")
cat("SD: ", round(sd(data), 2), "\n")
cat("Range:", range(data), "\n")

list(
  original = data,
  ma = moving_avg,
  mean = round(mean(data), 4),
  sd = round(sd(data), 4),
  min = round(min(data), 4),
  max = round(max(data), 4)
)
