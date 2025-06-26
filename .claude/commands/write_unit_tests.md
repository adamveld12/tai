---
allowed-tools: Bash(make test:*), Bash(make bench), Bash(make lint), Bash(make vet)
description: Write unit tests for the provided Go code.
---

I need help writing unit tests for the following go code at this path:

"${ARGUMENTS}"

Write the unit tests using the Go testing package. Ensure to cover various edge cases and scenarios to validate the functionality of the code.
Provide clear and concise test cases that can be easily understood and maintained.

## Guidelines for writing effective unit tests

### Areas to focus on:

- Hard and subtle issues
- Race conditions
- Error condition handling
- Configuration edge cases
- Boundary conditions
- Core logic nd functionality of the code

### Some good practices to follow:

- Use mock objects or stubs where necessary to isolate the code under test.
- Use table-driven tests for better organization and readability.
- Ensure tests are independent and can be run in any order.
- Use descriptive names for test functions and variables.
- Include comments to explain complex test scenarios.
- Always print the expected vs actual results in assertions for clarity.
- Always run the tests with `make test` to ensure they work correctly before you finish.

### Things to avoid:

- Avoid Testing external dependencies or integrations.
- Avoid Overly complex test setups.
- Avoid Redundant or duplicate test cases.
- Avoid Testing trivial code (ie getters/setters, simple data structures, typical language features)./
- Never edit the code you are testing. You are only allowed to edit the unit test files.

## Workflow

1. Thoughtfully review the provided Go code to understand its functionality and identify key areas that would most benefit from tests.
   - Make sure to understand the context and purpose of the code.
   - Identify any dependencies or external factors that may influence the behavior of the code.
   - Look for any complex logic or algorithms that may require thorough testing.
   - Consider any potential edge cases or scenarios that may not be immediately obvious.
   - Review subtle issues that may arise, like race conditions, nil pointers, or unexpected input values.

2. Based on your review, think about the different scenarios and edge cases that should be tested following the guidelines above.
   - list out these scenarios and edge cases before you start writing the tests.
   - Think and critically assess the quality and effectiveness of your tests.
   - Identify any areas for improvement or additional scenarios that could be covered.
   - Deeply reflect and think hard on the overall testing process and consider how it could be enhanced in future iterations.

3. Write one unit test at a time in a clear and organized manner, following practices outlined and additionally best practices for Go testing.
   - Use table-driven tests where appropriate to cover multiple scenarios in a single test function.
   - Ensure each test is independent and can be run in any order.
   - Use descriptive names for test functions and variables to enhance readability.
   - Include comments to explain complex test scenarios for better understanding.
   - Print expected vs actual results in assertions for clarity.
   - Use mock objects or stubs where necessary to isolate the code under test.
   - Avoid testing external dependencies or integrations to keep tests focused and reliable.
   - Avoid overly complex test setups to maintain simplicity and ease of understanding.

4. Run the test after you have completed it, to ensure it works before writing the next one.
   - Make sure to run the tests with `make test` to ensure they work correctly.
   - If the test fails, debug and fix the issue before proceeding to the next test.

5. Continue this process iteratively until you have covered all identified scenarios and edge cases.
   - never edit the code you are testing. You are only allowed to edit the unit test files.
   - if you identify any new scenarios or edge cases during the testing process, add them to your list and write tests for them as well.
   - if you absolutely must make a change to the code you are testing, make note of it and move on to the next test.

6. Once you have completed writing the tests, provide a summary of the tests written
   - include the scenarios covered and any important considerations or assumptions made during the testing process.
   - Ensure the summary is clear and concise, providing a quick overview of the test coverage achieved.
   - Note the code coverage percentage.

7. Finally, think hard on the overall testing process and consider how it could be improved in future iterations. Make a short list of improvements.
   - Think about any challenges faced during the testing process and how they were overcome.
   - Consider any areas where the testing process could be streamlined or made more efficient.
   - Reflect on the effectiveness of the tests written and whether they adequately cover the functionality of the code.
   - Identify any lessons learned that could be applied to future testing efforts.
   - Ask if I want to save this to a file.
