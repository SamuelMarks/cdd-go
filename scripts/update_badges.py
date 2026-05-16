#!/usr/bin/env python3
import os
import re
import subprocess

def get_color(pct):
    if pct >= 90: return 'brightgreen'
    if pct >= 80: return 'green'
    if pct >= 70: return 'yellowgreen'
    if pct >= 60: return 'yellow'
    if pct >= 50: return 'orange'
    return 'red'

def main():
    readme_path = os.path.join(os.path.dirname(__file__), '..', 'README.md')
    if not os.path.exists(readme_path):
        return

    # Calculate Test Coverage
    try:
        subprocess.run(["go", "test", "-coverprofile=coverage.out", "./..."], capture_output=True, text=True)
        res = subprocess.run(["go", "tool", "cover", "-func=coverage.out"], capture_output=True, text=True)
        out = res.stdout + res.stderr
        m = re.search(r'total:\s+\(statements\)\s+([0-9.]+)%', out)
        test_cov = float(m.group(1)) if m else 0.0
    except Exception as e:
        print(f'Test coverage calculation failed: {e}')
        test_cov = 0.0

    # Calculate Doc Coverage
    try:
        doc_res = subprocess.run(["go", "run", "./scripts/doc_cover.go"], capture_output=True, text=True)
        doc_out = doc_res.stdout.strip()
        doc_m = re.search(r'([0-9.]+)%', doc_out)
        doc_cov = float(doc_m.group(1)) if doc_m else 0.0
    except Exception as e:
        print(f'Doc coverage calculation failed: {e}')
        doc_cov = 0.0

    # Format numbers properly
    test_cov_str = f"{int(test_cov)}" if test_cov.is_integer() else f"{test_cov:.1f}"
    doc_cov_str = f"{int(doc_cov)}" if doc_cov.is_integer() else f"{doc_cov:.1f}"

    test_color = get_color(test_cov)
    doc_color = get_color(doc_cov)

    with open(readme_path, 'r') as f:
        content = f.read()

    content = re.sub(
        r'\[\!\[Test Coverage\]\(https://img\.shields\.io/badge/test_coverage-[0-9.]+%25-[a-z]+\.svg\)\]\(#\)',
        f'[![Test Coverage](https://img.shields.io/badge/test_coverage-{test_cov_str}%25-{test_color}.svg)](#)',
        content
    )

    content = re.sub(
        r'\[\!\[Doc Coverage\]\(https://img\.shields\.io/badge/doc_coverage-[0-9.]+%25-[a-z]+\.svg\)\]\(#\)',
        f'[![Doc Coverage](https://img.shields.io/badge/doc_coverage-{doc_cov_str}%25-{doc_color}.svg)](#)',
        content
    )

    with open(readme_path, 'w') as f:
        f.write(content)

if __name__ == '__main__':
    main()
