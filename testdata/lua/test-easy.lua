function All(...)
    print("all:", ...)
    return ...
end

x = {All(1, 22, 3)}

print("x[3]", x[3])

s = 1
table = {
    x, a = 1,
    ["b"] = 2,
    3,
    r = {
        c = 4,
        d = {
            e = 5
        }
    },
    s,
    inc = function(x)
        return x + 1
    end
}

print("table.b", table.a)

table.a = 6

assert(table.a == 6, "table.a should be 6")

table[6] = 7

a = table.r.d

a.hello = function ()
    return "Hello, World!"
end

function table.r.Sub(x, y)
    return x - y
end

function table.r.AddToC(r, x)
    r.c = r.c + x
    return r.c
end

function table.r.AddToC(r, x)
    r.c = r.c + x
    return r.c
end

function table.r.d:AddToE(x)
    self.e = self.e + x
    return self.e
end

print(a.hello())
print(table.r.Sub(10, 5))
print(table.r.AddToC(table.r, 3))
print(table.r.d:AddToE(2))

a = 2
b = 4
c = 3

-- This is a Lua script with various constructs

while a < 10 do
    a = a + 1
end

repeat
    if b == 4 then
        break
    end
    b = b + 1
until b >= 4

for i = 1, 3, 2 do
    c = c + i
end

if a == b then
    c = c + 1
elseif a > b then
    c = c - 1
else
    c = c * 2
end

function add(a, b, c)
    return a + b + c
end

local function multiply(a, b, c)
    return a * (b + c)
end

local div = function(a, b)
    if b == 0 then
        return nil, "Division by zero"
    end
    return a / b
end


-- local t = {name = "Lua", version = 5.4, isAwesome = true}
-- local f, s, var = pairs(t)
-- print(f)     --> function: 0x...
-- print(s)     --> table: 0x...
-- print(var)
-- while true do
--   local k, v = f(s, var)
--   if k == nil then break end
--   print(k, v)
--   var = k
-- end
--

s = 1
:: label1 ::

if s < 4 then
    if s == 3 then
        goto label2
    end
    s = s + 1
else
    goto label2
end


goto label1

:: label2 ::

print(s)

for i = 1, 10 do
    if i == 7 then
        goto skip
    end
    print(i)
end

:: skip ::

function sum(...)
    local args = {...}
    local sum = 0
    for i = 1, 100 do
        if args[i] == nil then
            break
        end
        sum = sum + args[i]
    end
    return sum
end

print("Sum of 1, 2, 3, 4, 5:", sum(1, 2, 3, 4, 5))

function first(...)
    local x = ...
    return x
end

print("first of 1, 2, 3, 4, 5:", first(1, 2, 3, 4, 5))

function factorial(n)
    if n == 0 then
        return 1
    else
        return n * factorial(n - 1)
    end
end

function ads(a)
    for i = 1, 100 do
        if a == i then
            return i
        end
    end
end

print("Factorial 5:", factorial(5))

return 0