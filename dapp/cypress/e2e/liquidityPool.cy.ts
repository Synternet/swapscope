describe('Liquidity Pool', () => {
  it('shows default chart and liquidity pool item list', () => {
    cy.visit('/');
    waitForItemsLoaded(18, 35);
    cy.matchImageSnapshot('liquidity-pool');
  });

  it('filters by date', () => {
    cy.visit('/');
    cy.getByTestId('DateFilter').within(() => cy.get('button').eq(1).click().blur());
    waitForItemsLoaded(22, 42);
    cy.matchImageSnapshot('liquidity-pool-filter-date-24h');

    cy.getByTestId('DateFilter').within(() => cy.get('button').eq(2).click().blur());
    waitForItemsLoaded(23, 44);
    cy.matchImageSnapshot('liquidity-pool-filter-date-48h');
  });

  it('filters by liquidity add size', () => {
    cy.visit('/');
    cy.getByTestId('PoolSizeFilter').within(() => cy.get('[data-index="1"]').first().click());
    waitForItemsLoaded(12, 24);
    cy.matchImageSnapshot('liquidity-pool-filter-size');
  });

  it('filters by token pair', () => {
    cy.visit('/');
    cy.getByTestId('TokenPairFilter').within(() => cy.get('button').should('have.length', 5));
    cy.getByTestId('TokenPairFilter').within(() => cy.get('button').eq(1).click().blur());
    waitForItemsLoaded(3, 11);
    cy.matchImageSnapshot('liquidity-pool-filter-token-pair-link-weth');
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
