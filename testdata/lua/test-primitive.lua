print(table.a)
table.b = func "a"
table[1] = table[2] + table[3]

table = {
    a = 1,
    ["b"] = 2,
    3,
    {
        c = 4,
        d = {
            e = 5
        }
    }
}

print "test-primitive.lua"

local a = 1 + 3 -- this is a comment

:: label1 ::
local function b(a)
    a = a + 1
    return a
end

--[==[
    this is a
    multi-line comment
    too
]=]=] x = 0
] ==] x = 1
]===] x = 2 ]==]

a = b(a)

function c.d.e:f(a)
    a = a + 2
    return a
end

goto label1

break