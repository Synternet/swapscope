describe('Liquidity Pool', () => {
  it('shows chart and liquidity pool item list', () => {
    cy.visit('/');
    cy.waitUntil(() => cy.getByTestId('LiquidityPoolTable').within(() => cy.get('tbody tr').should('have.length', 22)));
    cy.waitUntil(() =>
      cy.getByTestId('LiquidityPoolChart').within(() => cy.get('.plot .point').should('have.length', 42)),
    );
    cy.matchImageSnapshot('liquidity-pool');
  });
});

export {};
