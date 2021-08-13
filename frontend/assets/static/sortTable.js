// adapted from https://stackoverflow.com/a/57080195

document.querySelectorAll('table.sortable')
    .forEach((table)=> {
        table.querySelectorAll('th')
            .forEach((element, columnNo) => {
                element.addEventListener('click', event => {
                    if(element.classList.contains('ascSorted')) {
                        dir = -1;
                        element.classList.remove('ascSorted');
                        element.classList.add('descSorted');
                        element.innerText = element.innerText.slice(0,-2) + " ↓";
                    } else if(element.classList.contains('descSorted')) {
                        dir = 1;
                        element.classList.remove('descSorted');
                        element.classList.add('ascSorted');
                        element.innerText = element.innerText.slice(0,-2) + " ↑";
                    } else {
                        dir = 1;
                        element.classList.add('ascSorted');
                        element.innerText += " ↑";
                    }
                    sortTable(table, columnNo, 0, dir, 1);
                });
            });
    });

function sortTable(table, priCol, secCol, priDir, secDir) {
    const tableBody = table.querySelector('tbody');
    const tableData = table2data(tableBody);
    tableData.sort((a, b) => {
        if(a[priCol] === b[priCol]) {
            if(a[secCol] > b[secCol]) {
                return secDir;
            } else {
                return -secDir;
            }
        } else if(a[priCol] > b[priCol]) {
            return priDir;
        } else {
            return -priDir;
        }
    });
    data2table(tableBody, tableData);
}

function table2data(tableBody) {
    const tableData = [];
    tableBody.querySelectorAll('tr')
        .forEach(row => {
            const rowData = [];
            row.querySelectorAll('td')
                .forEach(cell => {
                    rowData.push(cell.innerHTML);
                });
            rowData.classList = row.classList.toString();
            tableData.push(rowData);
        });
    return tableData;
}

function data2table(tableBody, tableData) {
    tableBody.querySelectorAll('tr')
        .forEach((row, i) => {
            const rowData = tableData[i];
            row.classList = rowData.classList;
            row.querySelectorAll('td')
                .forEach((cell, j) => {
                    cell.innerHTML = rowData[j];
                });
            tableData.push(rowData);
        });
}
