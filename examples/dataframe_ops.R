# Data frame examples

ages <- 20:24
scores <- c(80, 85, 90, 75, 95)
df <- data.frame(age = ages, score = scores)
print(class(df))
print(nrow(df))
print(ncol(df))
print(df$score)

# add a new column
df$pass <- df$score > 80
print(head(df, 3))
