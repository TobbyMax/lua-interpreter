local a = 1 + 3 -- this is a comment

:: label1 ::
local function b(a)
    a = a + 1
    return a
end

function c.d.e:f(a)
    a = a + 2
    return a
end

goto label1

break