import { addMatchImageSnapshotCommand } from 'cypress-image-snapshot/command';
import 'cypress-wait-until';

addMatchImageSnapshotCommand({
  disableTimersAndAnimations: true,
});

Cypress.Commands.add('getByTestId', (testId: string) => {
  return cy.get(`[data-testid="${testId}"]`);
});