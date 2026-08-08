# STYLE_GUIDE.md — Code Style & Formatting Guidelines

## 1. Go Style Rules
- Standard `gofmt` and `goimports` formatting required on all source files.
- Exported functions, types, and constants MUST have clear Go docstrings starting with the symbol name.
- Package declarations MUST include a top-level package docstring.
- Explicit type declarations and error wrapping: `fmt.Errorf("failed to process item: %w", err)`.

## 2. SQL Style Rules
- Lowercase SQL keywords or standard UPPERCASE reserved words consistently.
- Explicit table and column constraints (`NOT NULL`, `PRIMARY KEY`, `FOREIGN KEY`).
