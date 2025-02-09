s = 1
table = {
    a = 1,
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

function table.r.d:AddToE(x)
    self.e = self.e + x
    return self.e
end

-- return a.hello()

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
return add(a, b, c)