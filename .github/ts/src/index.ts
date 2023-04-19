import { context } from "@actions/github";

const githubToken = process.env.GITHUB_TOKEN?.length;

console.log('Hello, world!' + githubToken)