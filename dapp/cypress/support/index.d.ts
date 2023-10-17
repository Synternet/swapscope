declare global {
  namespace Cypress {
    interface Chainable {
      getByTestId: (testId: string) => Chainable<JQuery<Element>>;
    }
  }
}

export {};
