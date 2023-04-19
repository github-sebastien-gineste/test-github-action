import * as fs from 'fs';
import * as path from 'path';

const file = fs.readFileSync(path.join(__dirname, './index.js'), {
    encoding: 'utf-8',
});
let subContent = file.split("// ----- Same -----")[1];
subContent = "module.exports = async ({github, context, core})=> {" + subContent + "}";

fs.writeFileSync(path.join(__dirname, './index.js'), subContent);