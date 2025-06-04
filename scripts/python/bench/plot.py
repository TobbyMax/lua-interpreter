import pandas as pd
import matplotlib.pyplot as plt

# Загрузить данные из CSV
csv_filename = "benchmark_results.csv"
df = pd.read_csv(csv_filename)

# Получить список интерпретаторов
interpreters = df["interpreter"].unique()

# === График 1: Время выполнения ===
plt.figure(figsize=(10, 6))
for interp in interpreters:
    subset = df[df["interpreter"] == interp]
    plt.plot(subset["n"], subset["time_ms"], marker='o', label=interp)

plt.title("Зависимость времени интерпретации от n")
plt.xlabel("n")
plt.ylabel("Время (мс)")
plt.grid(True)
plt.legend()
plt.tight_layout()
plt.savefig("benchmark_time_plot.png")
plt.show()

# === График 2: Пиковое использование памяти ===
plt.figure(figsize=(10, 6))
for interp in interpreters:
    subset = df[df["interpreter"] == interp]
    plt.plot(subset["n"], subset["peak_kb"], marker='o', label=interp)

plt.title("Зависимость пикового использования памяти от n")
plt.xlabel("n")
plt.ylabel("Пиковое потребление памяти (КБ)")
plt.grid(True)
plt.legend()
plt.tight_layout()
plt.savefig("benchmark_memory_plot.png")
plt.show()
