import { useEffect, useCallback } from 'react';
import { useAuthStore } from '../stores/authStore';
import { useWalletStore } from '../stores/walletStore';
import { useSSE } from '../hooks/useSSE';
import { WalletCard } from '../components/WalletCard';
import { TransactionList } from '../components/TransactionList';
import type { Transaction, Wallet } from '../types';

export function Dashboard() {
  const { user, logout } = useAuthStore();
  const {
    wallets,
    selectedWallet,
    transactions,
    isLoading,
    fetchWallets,
    fetchTransactions,
    selectWallet,
    updateWalletFromEvent,
    addTransactionFromEvent,
    updateTransactionFromEvent,
  } = useWalletStore();

  useEffect(() => {
    fetchWallets().catch(err => console.error('Failed to fetch wallets:', err));
  }, [fetchWallets]);

  useEffect(() => {
    if (selectedWallet) {
      fetchTransactions(selectedWallet.id).catch(err =>
        console.error('Failed to fetch transactions:', err)
      );
    }
  }, [selectedWallet, fetchTransactions]);

  // Handle SSE events
  const handleSSEEvent = useCallback(
    (event: { topic: string; event_type: string; data: Record<string, unknown> }) => {
      console.log('Received SSE event:', event);

      if (event.topic === 'wallets') {
        if (event.event_type === 'wallet.updated' || event.event_type === 'wallet.created') {
          updateWalletFromEvent(event.data as unknown as Partial<Wallet> & { id: string });
        }
      } else if (event.topic === 'transactions') {
        if (event.event_type === 'transaction.created') {
          addTransactionFromEvent(event.data as unknown as Transaction);
          // Also refetch wallets to update balance
          fetchWallets().catch(err => console.error('Failed to refetch wallets:', err));
        } else if (event.event_type === 'transaction.updated') {
          updateTransactionFromEvent(event.data as unknown as Partial<Transaction> & { id: string });
          // Also refetch wallets to update balance
          fetchWallets().catch(err => console.error('Failed to refetch wallets:', err));
        }
      }
    },
    [updateWalletFromEvent, addTransactionFromEvent, updateTransactionFromEvent, fetchWallets]
  );

  // Connect to SSE for real-time updates
  useSSE({
    topics: ['wallets', 'transactions'],
    onEvent: handleSSEEvent,
    onError: error => console.error('SSE error:', error),
    enabled: !!user,
  });

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Navigation */}
      <nav className="bg-white shadow-sm">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex justify-between h-16 items-center">
            <h1 className="text-xl font-bold text-primary-600">Nivo</h1>
            <div className="flex items-center space-x-4">
              <span className="text-gray-700">{user?.full_name}</span>
              <button onClick={logout} className="btn-secondary">
                Logout
              </button>
            </div>
          </div>
        </div>
      </nav>

      {/* Main Content */}
      <main className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {/* Welcome Section */}
        <div className="mb-8">
          <h2 className="text-3xl font-bold text-gray-900">Welcome back, {user?.full_name}!</h2>
          <p className="text-gray-600 mt-1">Here's your financial overview</p>
        </div>

        {/* Loading State */}
        {isLoading && wallets.length === 0 && (
          <div className="card text-center py-12">
            <div className="text-gray-500">Loading your wallets...</div>
          </div>
        )}

        {/* Wallets Section */}
        {!isLoading && wallets.length === 0 && (
          <div className="card text-center py-12">
            <div className="text-gray-500">
              <p className="text-lg">No wallets found</p>
              <p className="text-sm mt-2">Contact support to create your first wallet</p>
            </div>
          </div>
        )}

        {wallets.length > 0 && (
          <>
            {/* Wallets Grid */}
            <div className="mb-8">
              <h3 className="text-xl font-semibold text-gray-900 mb-4">Your Wallets</h3>
              <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
                {wallets.map(wallet => (
                  <WalletCard
                    key={wallet.id}
                    wallet={wallet}
                    isSelected={selectedWallet?.id === wallet.id}
                    onClick={() => selectWallet(wallet.id)}
                  />
                ))}
              </div>
            </div>

            {/* Quick Actions */}
            <div className="mb-8">
              <h3 className="text-xl font-semibold text-gray-900 mb-4">Quick Actions</h3>
              <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                <button className="btn-primary py-4 text-lg">
                  💸 Send Money
                </button>
                <button className="btn-primary py-4 text-lg">
                  💰 Deposit
                </button>
                <button className="btn-primary py-4 text-lg">
                  🏧 Withdraw
                </button>
              </div>
            </div>

            {/* Transactions Section */}
            {selectedWallet && (
              <div className="mb-8">
                <h3 className="text-xl font-semibold text-gray-900 mb-4">
                  Transactions - {selectedWallet.type.charAt(0).toUpperCase() + selectedWallet.type.slice(1)} Wallet
                </h3>
                {isLoading ? (
                  <div className="card text-center py-8 text-gray-500">
                    Loading transactions...
                  </div>
                ) : (
                  <TransactionList transactions={transactions} walletId={selectedWallet.id} />
                )}
              </div>
            )}
          </>
        )}
      </main>
    </div>
  );
}
