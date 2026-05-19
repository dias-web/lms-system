-- +goose Up
-- +goose StatementBegin

-- Courses --------------------------------------------------------------------
INSERT INTO courses (id, name, description) VALUES
(1, 'Golang Developer',
'Become a professional Golang developer. This course covers the language fundamentals, the standard library, concurrency primitives, building HTTP services with the net/http package and the Gin framework, working with relational databases through database/sql and GORM, writing idiomatic tests, and shipping production-ready services packaged with Docker. By the end of the course you will be comfortable designing, implementing and operating microservices written in Go.'),
(2, 'Python Developer',
'Become a professional Python developer. Master the language syntax, the standard library, popular web frameworks such as Django and FastAPI, ORM tools like SQLAlchemy, asynchronous programming with asyncio, testing with pytest, and packaging applications for deployment. The course is project-driven and ends with the construction of a small but realistic backend service.');

SELECT setval(pg_get_serial_sequence('courses', 'id'), (SELECT MAX(id) FROM courses));

-- Chapters for "Golang Developer" --------------------------------------------
INSERT INTO chapters (id, name, description, "order", course_id) VALUES
(1, 'Introduction to Go',
 'Get to know the Go programming language: its history, the design goals behind it, and what kinds of problems it is particularly well suited for. Install the toolchain, write your first program, and learn how Go projects are organised.',
 1, 1),
(2, 'Variables and Basic Types',
 'Learn how to declare variables, work with the built-in numeric and textual types, understand constants and zero values, and how Go infers types at compile time.',
 2, 1),
(3, 'Control Structures',
 'Branching and looping in Go. This chapter covers if/else statements, switch expressions, and the for loop in all of its forms. You will also learn how Go handles errors as values returned from functions and how that influences the way control flow is written in idiomatic Go code.',
 3, 1),
(4, 'Functions and Methods',
 'Defining functions, multiple return values, named results, variadic parameters, closures, and attaching methods to user-defined types. The chapter ends with a discussion of when to prefer functions over methods.',
 4, 1);

SELECT setval(pg_get_serial_sequence('chapters', 'id'), (SELECT MAX(id) FROM chapters));

-- Lessons for "Control Structures" -------------------------------------------
INSERT INTO lessons (id, name, description, content, "order", chapter_id) VALUES
(1, 'If-else Statement in Golang',
 'Learn how the if statement works in Go, why parentheses around the condition are not used, and how the optional initialisation form lets you scope helper variables to the branch where they are needed.',
 E'## The if statement\n\nGo''s if statements are like those in C, C++, JavaScript, Java, and Swift, except that the parentheses around the condition are not required and the braces around the body are mandatory.\n\n```go\nif x > 0 {\n    return math.Sqrt(x)\n}\n```\n\nThe expression after `if` must evaluate to a boolean value. Unlike many other languages, Go does NOT perform implicit conversion from numbers or pointers to booleans — you must write the comparison explicitly.\n\n## if with a short statement\n\nLike `for`, the `if` statement can start with a short statement to execute before the condition. Variables declared by the statement are only in scope until the end of the if/else block:\n\n```go\nif v := math.Pow(x, n); v < lim {\n    return v\n}\n```\n\nThis short-statement form is one of the most idiomatic patterns in Go: it lets you call a function that returns a value and an error, then branch on the error without leaking the temporary into the surrounding scope.\n\n## if/else\n\nVariables declared inside an `if` short statement are also available inside any of the `else` blocks.\n\n```go\nif v := math.Pow(x, n); v < lim {\n    return v\n} else {\n    fmt.Printf("%g >= %g\\n", v, lim)\n}\nreturn lim\n```\n\nNote that in idiomatic Go, when the `if` block ends with a `return`, the trailing `else` is usually omitted — the code that would have been in the else block is simply written at the outer level. This keeps the happy path on the left margin and reduces nesting.\n\n## Multiple conditions\n\nUse `else if` to chain conditions:\n\n```go\nfunc classify(n int) string {\n    if n < 0 {\n        return "negative"\n    } else if n == 0 {\n        return "zero"\n    } else {\n        return "positive"\n    }\n}\n```\n\nWhen there are more than two or three branches, prefer a `switch` statement — it is shorter and clearer.',
 1, 3),
(2, 'Switch Statement in Golang',
 'Learn how Go''s switch statement differs from C and Java: cases do not fall through by default, the switch can have no condition (acting as a clean if/else chain), and any type that supports equality can be used.',
 E'## The switch statement\n\nA `switch` statement is a shorter way to write a sequence of `if - else` statements. It runs the first case whose value is equal to the condition expression.\n\n```go\nswitch os := runtime.GOOS; os {\ncase "darwin":\n    fmt.Println("macOS.")\ncase "linux":\n    fmt.Println("Linux.")\ndefault:\n    fmt.Printf("%s.\\n", os)\n}\n```\n\n## No automatic fall-through\n\nGo''s switch cases need not be constants, and the values involved need not be integers. Cases are evaluated from top to bottom, stopping when a case succeeds. **Go automatically breaks at the end of each case**, so you do not need explicit `break` statements. Use the `fallthrough` keyword if you really want the next case to execute as well.\n\n## Switch with no condition\n\nSwitch without a condition is the same as `switch true`. This construct can be a clean way to write long if-then-else chains.\n\n```go\nt := time.Now()\nswitch {\ncase t.Hour() < 12:\n    fmt.Println("Good morning!")\ncase t.Hour() < 17:\n    fmt.Println("Good afternoon.")\ndefault:\n    fmt.Println("Good evening.")\n}\n```\n\nThis form is widely used in real Go code, often in place of long if/else if chains.',
 2, 3),
(3, 'For Loops in Golang',
 'Go has only one looping construct, the for loop. Learn its three forms: the classic C-style loop, the "while" form, and the infinite loop, plus the range-based variant for iterating collections.',
 E'## The for loop\n\nGo has only one looping construct, the `for` loop. The basic `for` loop has three components separated by semicolons:\n\n- the init statement: executed before the first iteration\n- the condition expression: evaluated before every iteration\n- the post statement: executed at the end of every iteration\n\n```go\nsum := 0\nfor i := 0; i < 10; i++ {\n    sum += i\n}\n```\n\nThe init and post statements are optional:\n\n```go\nsum := 1\nfor ; sum < 1000; {\n    sum += sum\n}\n```\n\nWhen only the condition is present, the semicolons can be dropped — and the result looks like a `while` loop from other languages:\n\n```go\nsum := 1\nfor sum < 1000 {\n    sum += sum\n}\n```\n\n## Infinite loop\n\nIf you omit the condition entirely, the loop runs forever. This is the canonical way to spell an infinite loop in Go:\n\n```go\nfor {\n    // do something forever\n}\n```\n\n## Range\n\nThe `range` form of the `for` loop iterates over a slice, array, map, string, or channel. On each iteration of a slice or array it returns two values: the index, and a copy of the element at that index.\n\n```go\nnums := []int{2, 3, 5, 7, 11, 13}\nfor i, value := range nums {\n    fmt.Printf("index %d, value %d\\n", i, value)\n}\n```\n\nIf you only need the index, drop the second variable. If you only need the value, use the blank identifier `_` for the index.',
 3, 3);

SELECT setval(pg_get_serial_sequence('lessons', 'id'), (SELECT MAX(id) FROM lessons));

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DELETE FROM lessons WHERE id IN (1, 2, 3);
DELETE FROM chapters WHERE id IN (1, 2, 3, 4);
DELETE FROM courses WHERE id IN (1, 2);
-- +goose StatementEnd