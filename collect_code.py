import os
import sys

# --- Configuration ---

# 1. The root directory of your project.
#    os.getcwd() assumes you run the script from your project's root folder.
PROJECT_ROOT = os.getcwd()

# 2. The name of the file to save all the code into.
OUTPUT_FILE = "output.md"

# 3. File extensions to look for.
FILE_EXTENSIONS = (".py", ".go", ".json", ".yaml", ".toml")

# 4. Directories to exclude.
#    This is crucial to avoid including thousands of files from node_modules.
EXCLUDE_DIRS = {
    "node_modules",
    ".git",
    ".idea",
    ".pytest_cache",
    ".venv",
    "tests",
    ".docs",
    ".__pycache__",
    ".expo",
    "assets",
    "web-build",
    "dist",
    "build",
    # Add any other directories you want to ignore here
}

# --- End of Configuration ---


def is_excluded(path, root_path):
    """Check if a path is in one of the excluded directories."""
    relative_path = os.path.relpath(path, root_path)
    parts = relative_path.split(os.path.sep)
    return any(part in EXCLUDE_DIRS for part in parts)


def collect_files():
    """Walks through the project, finds relevant files, and writes them to the output file."""
    file_count = 0

    # Open the output file in write mode, which will overwrite it if it exists.
    # We use utf-8 encoding for compatibility with all source code characters.
    try:
        with open(OUTPUT_FILE, "w", encoding="utf-8") as outfile:
            print(f"Starting file collection in '{PROJECT_ROOT}'...")
            print(f"Output will be saved to '{OUTPUT_FILE}'")

            # os.walk is the perfect tool for traversing a directory tree.
            for root, dirs, files in os.walk(PROJECT_ROOT, topdown=True):
                # Efficiently skip excluded directories
                dirs[:] = [d for d in dirs if d not in EXCLUDE_DIRS]

                for filename in files:
                    if filename.endswith(FILE_EXTENSIONS):
                        file_path = os.path.join(root, filename)

                        try:
                            with open(
                                file_path, "r", encoding="utf-8", errors="ignore"
                            ) as infile:
                                content = infile.read()

                                # Get a clean, relative path for the header
                                relative_path = os.path.relpath(file_path, PROJECT_ROOT)

                                # Use forward slashes for cross-platform consistency in the output
                                header_path = relative_path.replace(os.path.sep, "/")

                                print(f"  -> Adding {header_path}")

                                # Write the formatted content to the output file
                                outfile.write(f"--- \n\n")
                                outfile.write(f"**File:** `{header_path}`\n\n")
                                outfile.write("```typescript\n")
                                outfile.write(content)
                                outfile.write("\n```\n\n")

                                file_count += 1
                        except Exception as e:
                            print(f"    [!] Error reading file {file_path}: {e}")

    except IOError as e:
        print(f"Error: Could not write to output file {OUTPUT_FILE}: {e}")
        sys.exit(1)

    return file_count


if __name__ == "__main__":
    total_files = collect_files()
    if total_files > 0:
        print(f"\n✅ Success! Combined {total_files} files into '{OUTPUT_FILE}'.")
    else:
        print("\n⚠️ Warning: No files with the specified extensions were found.")
