# ─────────────────────────────────────────────
# smallR – New Features Showcase
# ─────────────────────────────────────────────

# === Constants ===
cat("π =", pi, "\n")
cat("T =", T, ", F =", F, "\n")
cat("First 5 letters:", letters[1:5], "\n")
cat("First 5 LETTERS:", LETTERS[1:5], "\n")

# === Math Functions ===
cat("\n--- Math ---\n")
cat("abs(-7)    =", abs(-7), "\n")
cat("sqrt(25)   =", sqrt(25), "\n")
cat("floor(3.7) =", floor(3.7), "\n")
cat("ceiling(3.2)=", ceiling(3.2), "\n")
cat("round(pi,3)=", round(pi, 3), "\n")
cat("exp(1)     =", round(exp(1), 4), "\n")
cat("log(exp(1))=", log(exp(1)), "\n")
cat("sign(-3)   =", sign(-3), "\n")

# Trigonometry
cat("sin(pi/2)  =", round(sin(pi / 2), 4), "\n")
cat("cos(0)     =", cos(0), "\n")

# Cumulative functions
x <- c(1, 2, 3, 4, 5)
cat("cumsum:    ", cumsum(x), "\n")
cat("cumprod:   ", cumprod(x), "\n")
cat("cummax:    ", cummax(c(1, 3, 2, 5, 4)), "\n")
cat("cummin:    ", cummin(c(5, 3, 4, 1, 2)), "\n")

cat("prod(1:5)  =", prod(1:5), "\n")
cat("diff(c(1,3,6,10)) =", diff(c(1, 3, 6, 10)), "\n")
cat("range(3,1,5)=", range(c(3, 1, 5)), "\n")

# === String Functions ===
cat("\n--- Strings ---\n")
cat(paste("Hello", "World"), "\n")
cat(paste0("No", "Space"), "\n")
cat(paste("a", "b", "c", sep="-"), "\n")
cat("nchar('hello')  =", nchar("hello"), "\n")
cat("substr('abcdef',2,4) =", substr("abcdef", 2, 4), "\n")
cat("toupper('hello') =", toupper("hello"), "\n")
cat("tolower('WORLD') =", tolower("WORLD"), "\n")
cat("trimws('  hi  ') =", trimws("  hi  "), "\n")

cat("startsWith('hello','hel')=", startsWith("hello", "hel"), "\n")
cat("endsWith('hello','llo')  =", endsWith("hello", "llo"), "\n")

cat("gsub('o','0','hello')    =", gsub("o", "0", "hello"), "\n")
cat("strrep('ab', 3)          =", strrep("ab", 3), "\n")
cat("sprintf('%s is %d', 'pi', 3) =", sprintf("%s is %d", "pi", 3), "\n")

# grepl returns logical vector
cat("grepl('lo', c('hello','world')) =", grepl("lo", c("hello", "world")), "\n")

# strsplit
parts <- strsplit("a,b,c", ",")
cat("strsplit('a,b,c', ',')[[1]] =", parts[[1]], "\n")

# === Vector Utilities ===
cat("\n--- Vector Utilities ---\n")
v <- c(3, 1, 4, 1, 5, 9, 2, 6)
cat("sort:      ", sort(v), "\n")
cat("sort(dec): ", sort(v, decreasing=TRUE), "\n")
cat("rev:       ", rev(v), "\n")
cat("unique:    ", unique(v), "\n")
cat("duplicated:", duplicated(v), "\n")
cat("order:     ", order(v), "\n")

cat("which(c(F,T,F,T))  =", which(c(FALSE, TRUE, FALSE, TRUE)), "\n")
cat("which.min(c(3,1,2))=", which.min(c(3, 1, 2)), "\n")
cat("which.max(c(3,1,2))=", which.max(c(3, 1, 2)), "\n")

cat("any(c(F,T,F))=", any(c(FALSE, TRUE, FALSE)), "\n")
cat("all(c(T,T,T))=", all(c(TRUE, TRUE, TRUE)), "\n")

# Set operations
cat("union(1:3, 3:5)    =", union(1:3, 3:5), "\n")
cat("intersect(1:3, 2:4)=", intersect(1:3, 2:4), "\n")
cat("setdiff(1:3, 2:4)  =", setdiff(1:3, 2:4), "\n")

cat("match(c('b','d','a'), c('a','b','c')) =", match(c("b","d","a"), c("a","b","c")), "\n")
cat("seq_len(5)   =", seq_len(5), "\n")
cat("seq_along(v) =", seq_along(v), "\n")
cat("append(1:3, 4:5) =", append(1:3, 4:5), "\n")

cat("table(c('a','b','a','c','b','a')):\n")
print(table(c("a", "b", "a", "c", "b", "a")))
cat("tabulate(c(2,3,3,5)) =", tabulate(c(2, 3, 3, 5)), "\n")

# === Type Checking ===
cat("\n--- Type Checking ---\n")
cat("is.numeric(1)     =", is.numeric(1), "\n")
cat("is.character('a') =", is.character("a"), "\n")
cat("is.logical(TRUE)  =", is.logical(TRUE), "\n")
cat("is.null(NULL)     =", is.null(NULL), "\n")
cat("is.list(list(1))  =", is.list(list(1)), "\n")
cat("is.function(sum)  =", is.function(sum), "\n")
cat("is.finite(Inf)    =", is.finite(Inf), "\n")
cat("is.nan(NaN)       =", is.nan(NaN), "\n")
cat("is.infinite(Inf)  =", is.infinite(Inf), "\n")
cat("identical(1, 1)   =", identical(1, 1), "\n")

# === Apply Family ===
cat("\n--- Apply Family ---\n")
result <- sapply(c(1, 4, 9, 16), sqrt)
cat("sapply(c(1,4,9,16), sqrt) =", result, "\n")

doubled <- lapply(list(1, 2, 3), function(x) x * 2)
cat("lapply double: ")
for (i in 1:3) cat(doubled[[i]], "")
cat("\n")

total <- Reduce(function(a, b) a + b, 1:5, 0)
cat("Reduce(+, 1:5, 0) =", total, "\n")

big <- Filter(function(x) x > 3, c(1, 2, 3, 4, 5))
cat("Filter(>3, 1:5): ")
for (i in seq_along(big)) cat(big[[i]], "")
cat("\n")

cat("do.call(paste, list('x','y','z')) =", do.call(paste, list("x", "y", "z")), "\n")

# === Operators ===
cat("\n--- Operators ---\n")

# %in% operator
cat("2 %in% c(1,2,3) =", 2 %in% c(1, 2, 3), "\n")
cat("4 %in% c(1,2,3) =", 4 %in% c(1, 2, 3), "\n")
cat("c(1,4,2) %in% c(1,2,3) =", c(1, 4, 2) %in% c(1, 2, 3), "\n")

# |> pipe operator
piped <- c(5, 3, 1, 4, 2) |> sort() |> rev()
cat("c(5,3,1,4,2) |> sort() |> rev() =", piped, "\n")

total_pipe <- c(1, 2, 3, 4, 5) |> sum()
cat("1:5 |> sum() =", total_pipe, "\n")

# === Control Flow ===
cat("\n--- Control Flow ---\n")

# ifelse (vectorized)
cat("ifelse(c(T,F,T), 'yes', 'no') =", ifelse(c(TRUE, FALSE, TRUE), c("yes", "yes", "yes"), c("no", "no", "no")), "\n")

# switch
result <- switch("b", a = 1, b = 2, c = 3)
cat("switch('b', a=1, b=2, c=3) =", result, "\n")

# tryCatch
safe <- tryCatch(stop("oops"), error = function(e) paste("caught:", e))
cat("tryCatch(stop('oops')) =", safe, "\n")

# exists
myvar <- 42
cat("exists('myvar') =", exists("myvar"), "\n")
cat("exists('nope')  =", exists("nope"), "\n")

cat("\nDone!\n")
