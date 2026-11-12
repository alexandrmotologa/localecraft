# ICU MessageFormat Validation Guide

`localecraft` includes an Abstract Syntax Tree (AST) parser and validator for ICU MessageFormat strings.

## Common Syntax Patterns

### Simple Placeholders
```text
Hello, {userName}! You have {unreadCount} unread notifications.
```
Rules:
- Placeholders in target translations must match source placeholders exactly.
- Renaming `{userName}` to `{user}` or omitting `{unreadCount}` triggers validation failure.

### Plural Selection
```text
{count, plural, =0{No files remaining} one{# file remaining} other{# files remaining}}
```
Rules:
- The `#` character inside a plural branch interpolates the formatted count value.
- Exact numeric cases like `=0` and `=1` are permitted across all locales.
- The `other` case is mandatory in all ICU plural selections.

### Locale-Specific CLDR Categories
Plural category compliance varies per language:
- English, German, French, Spanish: `one`, `other`
- Russian, Ukrainian, Polish: `one`, `few`, `many`, `other`
- Romanian: `one`, `few`, `other`
- Arabic: `zero`, `one`, `two`, `few`, `many`, `other`
- Japanese, Chinese: `other`

`localecraft` flags target translations that use invalid plural categories for the target locale.

### Gender and Select Variants
```text
{gender, select, male{He updated his profile} female{She updated her profile} other{They updated their profile}}
```
Rules:
- Requires a fallback `other` branch.
