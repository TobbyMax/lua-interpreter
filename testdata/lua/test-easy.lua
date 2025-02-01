a = 1
b = 4
c = 3

while a < 10 do
    a = a + 1
end

repeat
    b = b + 1
until b >= 4

for i = 1, 3, 2 do
    c = c + i
end

if a > b then
    c = c + 1
elseif a < b then
    c = c - 1
else
    c = c * 2
end

return a + b + c