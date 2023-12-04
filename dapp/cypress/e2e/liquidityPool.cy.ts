describe('Liquidity Pool', () => {
  it('shows default chart and liquidity pool item list', () => {
    cy.visit('/');
    waitForItemsLoaded(7, 14);
    cy.matchImageSnapshot('liquidity-pool');
  });

  it('filters by date', () => {
    cy.visit('/');
    cy.getByTestId('DateFilter').within(() => cy.get('button').eq(1).click().blur());
    waitForItemsLoaded(15, 30);
    cy.matchImageSnapshot('liquidity-pool-filter-date-24h');

    cy.getByTestId('DateFilter').within(() => cy.get('button').eq(2).click().blur());
    waitForItemsLoaded(15, 30);
    cy.matchImageSnapshot('liquidity-pool-filter-date-48h');
  });

  it('filters by liquidity value', () => {
    cy.visit('/');
    cy.getByTestId('LiquidityValueFilter').within(() => cy.get('[data-index="1"]').first().click());
    waitForItemsLoaded(4, 8);
    cy.matchImageSnapshot('liquidity-pool-filter-size');
  });

  it('filters by token pair', () => {
    cy.visit('/');
    cy.getByTestId('TokenPairFilter').within(() => cy.get('button').should('have.length', 5));
    cy.getByTestId('TokenPairFilter').within(() => cy.get('button').eq(1).click().blur());
    waitForItemsLoaded(1, 2);
    cy.matchImageSnapshot('liquidity-pool-filter-token-pair-link-weth');
  });

  it('filters by operation type', () => {
    cy.visit('/');
    cy.getByTestId('OperationTypeFilter').click();
    cy.getByTestId('OperationTypeFilter-add').click();
    waitForItemsLoaded(3, 6);
    cy.matchImageSnapshot('liquidity-pool-filter-operation-type-add');

    cy.getByTestId('OperationTypeFilter').click();
    cy.getByTestId('OperationTypeFilter-remove').click();
    waitForItemsLoaded(4, 8);
    cy.matchImageSnapshot('liquidity-pool-filter-operation-type-remove');
  });
  

  it.only('zooms in chart', () => {
    cy.visit('/');
    cy.getByTestId('DateFilter').within(() => cy.get('button').eq(1).click().blur());
    waitForItemsLoaded(15, 30);
    cy.getByTestId('LiquidityPoolChart').within(() => cy.get('.modebar a[data-title="Zoom in"]').click());
    cy.matchImageSnapshot('liquidity-pool-zoom-in');
  });
});

function waitForItemsLoaded(tableRowCount: number, chartPointCount: number) {
  cy.waitUntil(() =>
    cy.getByTestId('LiquidityPoolTable').within(() => cy.get('tbody tr').should('have.length', tableRowCount)),
  );
  cy.waitUntil(() =>
    cy.getByTestId('LiquidityPoolChart').within(() => cy.get('.plot .point').should('have.length', chartPointCount)),
  );
}

export {};
