print "test-primitive.lua"
-- table = {
--     a = 1,
--     ["b"] = 2,
--     3
-- }
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