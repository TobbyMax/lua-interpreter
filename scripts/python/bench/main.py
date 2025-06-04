import os
import time
import subprocess
import csv
from typing import Tuple

import psutil

# Пути к интерпретаторам
LUA_EXECUTABLE = r"C:\Dev\lua-5.4.2_Win64_bin\lua.exe"
GUA_EXECUTABLE = r"C:\Dev\lua-interpreter\gua.exe"
LUAGO_EXECUTABLE = r"C:\Dev\LuaInterpreter\src\luago\luago.exe"
GO_LUA_EXECUTABLE = r"C:\Dev\test-go-lua\go-lua.exe"
GLUA_EXECUTABLE = r"C:\Dev\gopher-lua\glua.exe"

# Файл с результатами
CSV_FILENAME = "result/benchmark_results_recursive.csv"

# Количество повторений
NUM_RUNS = 5


def generate_lua_script(n_value: int, filename: str):
    lua_code = f"""\
function fibt(n0, n1, c)
    if c == 0 then
        return n0
    else if c == 1 then
        return n1
    end
    return fibt(n1, n0+n1, c-1)
end
end

function fib(n)
    return fibt(0, 1, n)
end

fib({n_value})
"""
    with open(filename, "w", encoding="utf-8") as f:
        f.write(lua_code)


def run_interpreter(executable: str, filename: str) -> Tuple[float, float]:
    proc = psutil.Popen([executable, filename], stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)

    peak_memory = 0  # В КБ
    start_time = time.perf_counter()

    while proc.is_running():
        try:
            mem = proc.memory_info().rss  # В байтах
            peak_memory = max(peak_memory, mem)
        except psutil.NoSuchProcess:
            break
        time.sleep(0.001)  # немного подождать, чтобы не перегрузить CPU

    proc.wait()
    end_time = time.perf_counter()

    elapsed_ms = (end_time - start_time) * 1000
    peak_kb = peak_memory / 1024  # В килобайтах
    return elapsed_ms, peak_kb


def average_time(executable: str, filename: str, runs: int) -> Tuple[float, float]:
    times = []
    mems = []
    for _ in range(runs):
        elapsed_ms, peak_kb = run_interpreter(executable, filename)
        times.append(elapsed_ms)
        mems.append(peak_kb)
    return sum(times) / runs, sum(mems) / runs


def benchmark(n_values, num_runs):
    results = []

    for n in n_values:
        lua_filename = f"script_{n}.lua"
        generate_lua_script(n, lua_filename)

        # Lua
        avg_time, avg_mem = average_time(LUA_EXECUTABLE, lua_filename, num_runs)
        print(f"[lua]   n={n:7d} | avg time = {avg_time:.3f} ms | peak memory = {avg_mem:.1f} KB")
        results.append(["lua", n, round(avg_time, 3), round(avg_mem, 1)])

        # Gua
        avg_time, avg_mem = average_time(GUA_EXECUTABLE, lua_filename, num_runs)
        print(f"[gua]   n={n:7d} | avg time = {avg_time:.3f} ms | peak memory = {avg_mem:.1f} KB")
        results.append(["gua", n, round(avg_time, 3), round(avg_mem, 1)])

        # Luago
        avg_luago, avg_mem = average_time(LUAGO_EXECUTABLE, lua_filename, num_runs)
        print(f"[luago] n={n:7d} | avg time = {avg_luago:.3f} ms | peak memory = {avg_mem:.1f} KB")
        results.append(["luago", n, round(avg_luago, 3), round(avg_mem, 1)])

        # Go Lua
        avg_go_lua, avg_mem = average_time(GO_LUA_EXECUTABLE, lua_filename, num_runs)
        print(f"[go-lua] n={n:7d} | avg time = {avg_go_lua:.3f} ms | peak memory = {avg_mem:.1f} KB")
        results.append(["go-lua", n, round(avg_go_lua, 3), round(avg_mem, 1)])

        # Glua
        avg_glua, avg_mem = average_time(GLUA_EXECUTABLE, lua_filename, num_runs)
        print(f"[glua]  n={n:7d} | avg time = {avg_glua:.3f} ms | peak memory = {avg_mem:.1f} KB")
        results.append(["glua", n, round(avg_glua, 3), round(avg_mem, 1)])

        os.remove(lua_filename)

    return results


def save_results_to_csv(results, filename):
    with open(filename, mode="w", newline="", encoding="utf-8") as csvfile:
        writer = csv.writer(csvfile)
        writer.writerow(["interpreter", "n", "time_ms", "peak_kb"])
        writer.writerows(results)
    print(f"\nРезультаты сохранены в {filename}")


# Значения n
n_values = [10, 50, 100, 500, 1000, 5000, 10000, 20000]

# Запуск и сохранение
results = benchmark(n_values, NUM_RUNS)
save_results_to_csv(results, CSV_FILENAME)
